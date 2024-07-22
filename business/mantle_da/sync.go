package mantle_da

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/google/uuid"

	"github.com/mantlenetworkio/lithosphere/database/business"
	mantle_da "github.com/mantlenetworkio/lithosphere/database/event/mantle-da"
	"github.com/mantlenetworkio/lithosphere/synchronizer/mantle-da/common/graphView"
)

const ConfirmDataStoreEventABI = "ConfirmDataStore(uint32,bytes32)"

var ConfirmDataStoreEventABIHash = crypto.Keccak256Hash([]byte(ConfirmDataStoreEventABI))

func ParseMantleDaEvent(DataLayrServiceManagerAddr string, logs []types.Log, log log.Logger) ([]mantle_da.DataStoreEvent, error) {
	abiUint32, err := abi.NewType("uint32", "uint32", nil)
	if err != nil {
		log.Error("Abi new uint32 type error", "err", err)
		return nil, err
	}
	abiBytes32, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		log.Error("Abi new bytes32 type error", "err", err)
		return nil, err
	}
	confirmDataStoreArgs := abi.Arguments{
		{
			Name:    "dataStoreId",
			Type:    abiUint32,
			Indexed: false,
		}, {
			Name:    "headerHash",
			Type:    abiBytes32,
			Indexed: false,
		},
	}
	var dataStoreEvents = []mantle_da.DataStoreEvent{}
	var dataStoreData = make(map[string]interface{})
	for _, rLog := range logs {
		if strings.ToLower(rLog.Address.String()) != strings.ToLower(DataLayrServiceManagerAddr) {
			continue
		}
		if rLog.Topics[0] != ConfirmDataStoreEventABIHash {
			continue
		}
		if len(rLog.Data) > 0 {
			err := confirmDataStoreArgs.UnpackIntoMap(dataStoreData, rLog.Data)
			if err != nil {
				log.Error("Unpack data into map fail", "err", err)
				continue
			}
			if dataStoreData != nil {
				log.Info("Parse confirmed dataStoreId success",
					"dataStoreId", dataStoreData["dataStoreId"].(uint32),
					"blockHash", rLog.BlockHash.String())
				dataStoreEvent := constructDataStoreEvent(dataStoreData["dataStoreId"].(uint32), rLog.BlockHash)
				dataStoreEvents = append(dataStoreEvents, dataStoreEvent)
			}
		}
	}
	return dataStoreEvents, nil
}

func DataFromMantleDa(deList []*mantle_da.DataStoreEvent, da *MantleDataStore, log log.Logger) ([]business.DataStore, []business.DataStoreBlock, uint32, error) {
	var dataStores = []business.DataStore{}
	var dataStoreBlock = []business.DataStoreBlock{}
	var latestDataStoreId uint32
	for _, dIdList := range deList {
		log.Info("Get data from mantle da", "DataStoreId", dIdList.DataStoreId)
		datastore, err := da.RetrievalDataStoreFromDa(uint32(dIdList.DataStoreId))
		if err != nil {
			log.Warn("Retrieval frames from mantleDa error", "dataStoreId", dIdList.DataStoreId, "err", err)
		}
		if datastore != nil {
			if !datastore.Confirmed {
				log.Warn("This batch is not confirmed")
			}
			frames, err := da.getFramesByDataStoreId(uint32(dIdList.DataStoreId))
			if err != nil {
				log.Warn("Get frames fail", "err", err)
			}
			cDataStoreBlock := constructDataStoreBlock(datastore, frames)
			dataStoreBlock = append(dataStoreBlock, cDataStoreBlock)
			cDataStore := constructDataStore(datastore)
			dataStores = append(dataStores, cDataStore)
		}
		latestDataStoreId = uint32(dIdList.DataStoreId)
	}
	return dataStores, dataStoreBlock, latestDataStoreId, nil
}

func constructDataStoreEvent(dataStoreId uint32, blockHash common.Hash) mantle_da.DataStoreEvent {
	dataStoreEvent := mantle_da.DataStoreEvent{
		GUID:        uuid.New(),
		DataStoreId: uint64(dataStoreId),
		BlockHash:   blockHash,
		Timestamp:   uint64(time.Now().Unix()),
	}
	return dataStoreEvent
}

func constructDataStore(ds *graphView.DataStore) business.DataStore {
	dataStore := business.DataStore{
		GUID:                 uuid.New(),
		DataStoreId:          uint64(ds.StoreNumber),
		DurationDataStoreId:  uint64(ds.DurationDataStoreId),
		DataSize:             new(big.Int).SetUint64(1), //default
		DataInitHash:         ds.InitTxHash,
		DataConfirmHash:      ds.ConfirmTxHash,
		FromStoreNumber:      new(big.Int).SetUint64(uint64(ds.ReferenceBlockNumber)),
		StakeFromBlockNumber: ds.InitBlockNumber,
		InitGasUsed:          new(big.Int).SetUint64(ds.InitGasUsed),
		InitBlockNumber:      ds.InitBlockNumber,
		ConfirmGasUsed:       new(big.Int).SetUint64(ds.ConfirmGasUsed),
		DataHash:             hex.EncodeToString(ds.MsgHash[:]),
		EthSign:              ds.EthSigned.String(),
		MantleSign:           ds.EigenSigned.String(),
		SignatoryRecord:      hex.EncodeToString(ds.SignatoryRecord[:]),
		InitTime:             uint64(ds.InitTime),
		ExpireTime:           uint64(ds.ExpireTime),
		NumSys:               ds.NumSys,
		NumPar:               ds.NumPar,
		LowDegree:            strconv.Itoa(int(ds.Degree)),
		Status:               ds.Confirmed,
		Confirmer:            ds.Confirmer,
		Header:               hex.EncodeToString(ds.Header),
		DataCommitment:       hex.EncodeToString(ds.DataCommitment[:]),
		Timestamp:            uint64(time.Now().Unix()),
	}
	return dataStore
}

func constructDataStoreBlock(ds *graphView.DataStore, frames []byte) business.DataStoreBlock {
	dataStoreBlock := business.DataStoreBlock{
		GUID:          uuid.New(),
		DataStoreID:   uint64(ds.StoreNumber),
		BlockData:     common.HexToHash(hex.EncodeToString(frames)),
		L2TxHash:      common.HexToHash(""),
		L2BlockNumber: nil,
		Canonical:     true,
		Timestamp:     uint64(time.Now().Unix()),
	}
	return dataStoreBlock
}
