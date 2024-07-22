// Database module defines the data DB struct which wraps specific DB interfaces for L1/L2 block headers, contract events, bridging schemas.
package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/log"

	"github.com/mantlenetworkio/lithosphere/config"
	"github.com/mantlenetworkio/lithosphere/database/business"
	"github.com/mantlenetworkio/lithosphere/database/common"
	"github.com/mantlenetworkio/lithosphere/database/event"
	mantle_da "github.com/mantlenetworkio/lithosphere/database/event/mantle-da"
	"github.com/mantlenetworkio/lithosphere/database/exporter"
	"github.com/mantlenetworkio/lithosphere/database/utils"
	_ "github.com/mantlenetworkio/lithosphere/database/utils/serializers"
	v1 "github.com/mantlenetworkio/lithosphere/database/v1"
	"github.com/mantlenetworkio/lithosphere/synchronizer/retry"
)

type DB struct {
	gorm *gorm.DB

	Blocks             common.BlocksDB
	Transactions       common.TransactionsDB
	ContractEvents     event.ContractEventsDB
	WithdrawProven     event.WithdrawProvenDB
	WithdrawFinalized  event.WithdrawFinalizedDB
	RelayMessage       event.RelayMessageDB
	StateRoots         business.StateRootDB
	DataStore          business.DataStoreDB
	L2ToL1             business.L2ToL1DB
	L1ToL2             business.L1ToL2DB
	DataStoreEvent     mantle_da.DataStoreEventDB
	L2SentMessageEvent v1.L2SentMessageEventDB
	CheckPoint         exporter.BridgeCheckpointDB
	TokenList          business.TokenListDB
}

func NewDB(ctx context.Context, log log.Logger, dbConfig config.DBConfig) (*DB, error) {
	log = log.New("module", "db")

	dsn := fmt.Sprintf("host=%s dbname=%s sslmode=disable", dbConfig.Host, dbConfig.Name)
	if dbConfig.Port != 0 {
		dsn += fmt.Sprintf(" port=%d", dbConfig.Port)
	}
	if dbConfig.User != "" {
		dsn += fmt.Sprintf(" user=%s", dbConfig.User)
	}
	if dbConfig.Password != "" {
		dsn += fmt.Sprintf(" password=%s", dbConfig.Password)
	}

	gormConfig := gorm.Config{
		Logger:                 utils.NewLogger(log),
		SkipDefaultTransaction: true,
		CreateBatchSize:        3_000,
	}

	retryStrategy := &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
	gorm, err := retry.Do[*gorm.DB](context.Background(), 10, retryStrategy, func() (*gorm.DB, error) {
		gorm, err := gorm.Open(postgres.Open(dsn), &gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		return gorm, nil
	})

	if err != nil {
		return nil, err
	}

	db := &DB{
		gorm:               gorm,
		Blocks:             common.NewBlocksDB(gorm),
		Transactions:       common.NewTransactionsDB(gorm),
		ContractEvents:     event.NewContractEventsDB(gorm),
		WithdrawProven:     event.NewWithdrawProvenDB(gorm),
		WithdrawFinalized:  event.NewWithdrawFinalizedDB(gorm),
		RelayMessage:       event.NewRelayMessageDB(gorm),
		StateRoots:         business.NewStateRootDB(gorm),
		DataStore:          business.NewDataStoreDB(gorm),
		L1ToL2:             business.NewL1ToL2DB(gorm),
		L2ToL1:             business.NewL21ToL1DB(gorm),
		DataStoreEvent:     mantle_da.NewDataStoreEvnetDB(gorm),
		L2SentMessageEvent: v1.NewL2SentMessageEventDB(gorm),
		CheckPoint:         exporter.NewBridgeCheckpointDB(gorm),
		TokenList:          business.NewTokenListDB(gorm),
	}
	return db, nil
}

// Transaction executes all operations conducted with the supplied database in a single
// transaction. If the supplied function errors, the transaction is rolled back.
func (db *DB) Transaction(fn func(db *DB) error) error {
	return db.gorm.Transaction(func(tx *gorm.DB) error {
		txDB := &DB{
			gorm:               tx,
			Blocks:             common.NewBlocksDB(tx),
			Transactions:       common.NewTransactionsDB(tx),
			ContractEvents:     event.NewContractEventsDB(tx),
			WithdrawProven:     event.NewWithdrawProvenDB(tx),
			WithdrawFinalized:  event.NewWithdrawFinalizedDB(tx),
			RelayMessage:       event.NewRelayMessageDB(tx),
			DataStore:          business.NewDataStoreDB(tx),
			L1ToL2:             business.NewL1ToL2DB(tx),
			L2ToL1:             business.NewL21ToL1DB(tx),
			StateRoots:         business.NewStateRootDB(tx),
			DataStoreEvent:     mantle_da.NewDataStoreEvnetDB(tx),
			L2SentMessageEvent: v1.NewL2SentMessageEventDB(tx),
			CheckPoint:         exporter.NewBridgeCheckpointDB(tx),
			TokenList:          business.NewTokenListDB(tx),
		}
		return fn(txDB)
	})
}

func (db *DB) Close() error {
	sql, err := db.gorm.DB()
	if err != nil {
		return err
	}
	return sql.Close()
}

func (db *DB) ExecuteSQLMigration(migrationsFolder string) error {
	err := filepath.Walk(migrationsFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to process migration file: %s", path))
		}
		if info.IsDir() {
			return nil
		}
		fileContent, readErr := os.ReadFile(path)
		if readErr != nil {
			return errors.Wrap(readErr, fmt.Sprintf("Error reading SQL file: %s", path))
		}

		execErr := db.gorm.Exec(string(fileContent)).Error
		if execErr != nil {
			return errors.Wrap(execErr, fmt.Sprintf("Error executing SQL script: %s", path))
		}
		return nil
	})
	return err
}
