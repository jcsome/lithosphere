package mantle_da

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ethereum/go-ethereum/log"

	"github.com/mantlenetworkio/lithosphere/config"
	"github.com/mantlenetworkio/lithosphere/synchronizer/mantle-da/common/graphView"
	pb "github.com/mantlenetworkio/lithosphere/synchronizer/mantle-da/common/interfaces/interfaceRetrieverServer"
)

const (
	POLLING_INTERVAL     = 1 * time.Second
	MAX_RPC_MESSAGE_SIZE = 1024 * 1024 * 300
)

type MantleDataStoreConfig struct {
	RetrieverSocket          string
	RetrieverTimeout         time.Duration
	GraphProvider            string
	DataStorePollingDuration time.Duration
}

type MantleDataStore struct {
	Ctx           context.Context
	Cfg           *MantleDataStoreConfig
	GraphClient   *graphView.GraphClient
	GraphqlClient *graphql.Client
}

func NewMantleDataStore(cfg *MantleDataStoreConfig) (*MantleDataStore, error) {
	ctx := context.Background()
	graphClient := graphView.NewGraphClient(cfg.GraphProvider, nil)
	graphqlClient := graphql.NewClient(graphClient.GetEndpoint(), nil)
	mDatastore := &MantleDataStore{
		Ctx:           ctx,
		Cfg:           cfg,
		GraphClient:   graphClient,
		GraphqlClient: graphqlClient,
	}
	return mDatastore, nil
}

func NewMantleDataStoreConfig(config config.DAConfig) (MantleDataStoreConfig, error) {
	return MantleDataStoreConfig{
		RetrieverSocket:          config.RetrieverSocket,
		RetrieverTimeout:         config.RetrieverTimeout,
		GraphProvider:            config.GraphProvider,
		DataStorePollingDuration: config.DataStorePollingDuration,
	}, nil
}

func (mda *MantleDataStore) getDataStoreById(dataStoreId uint32) (*graphView.DataStore, error) {
	var query struct {
		DataStore graphView.DataStoreGql `graphql:"dataStore(id: $storeId)"`
	}
	variables := map[string]interface{}{
		"storeId": graphql.String(strconv.FormatUint(uint64(dataStoreId), 10)),
	}
	err := mda.GraphqlClient.Query(mda.Ctx, &query, variables)
	if err != nil {
		return nil, err
	}
	log.Debug("Query dataStore success",
		"DurationDataStoreId", query.DataStore.DurationDataStoreId,
		"Confirmed", query.DataStore.Confirmed,
		"ConfirmTxHash", query.DataStore.ConfirmTxHash)
	dataStore, err := query.DataStore.Convert()
	if err != nil {
		log.Warn("DataStoreGql convert to DataStore fail", "err", err)
		return nil, err
	}
	return dataStore, nil
}

func (mda *MantleDataStore) getFramesByDataStoreId(dataStoreId uint32) ([]byte, error) {
	conn, err := grpc.Dial(mda.Cfg.RetrieverSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Connect to da retriever fail", "err", err)
		return nil, err
	}
	defer conn.Close()
	client := pb.NewDataRetrievalClient(conn)
	opt := grpc.MaxCallRecvMsgSize(MAX_RPC_MESSAGE_SIZE)
	request := &pb.FramesAndDataRequest{
		DataStoreId: dataStoreId,
	}
	reply, err := client.RetrieveFramesAndData(mda.Ctx, request, opt)
	if err != nil {
		log.Warn("Retrieve frames and data fail", "err", err)
		return nil, err
	}
	log.Debug("Get reply data success", "reply length", len(reply.GetData()))
	return reply.GetData(), nil
}

func (mda *MantleDataStore) RetrievalDataStoreFromDa(dataStoreId uint32) (*graphView.DataStore, error) {
	pollingTimeout := time.NewTimer(mda.Cfg.DataStorePollingDuration)
	defer pollingTimeout.Stop()
	intervalTicker := time.NewTicker(POLLING_INTERVAL)
	defer intervalTicker.Stop()
	for {
		select {
		case <-intervalTicker.C:
			if dataStoreId <= 0 {
				log.Warn("DataStoreId less than zero", "dataStoreId", dataStoreId)
				return nil, errors.New("dataStoreId less than 0")
			}
			dataStore, err := mda.getDataStoreById(dataStoreId)
			if err != nil {
				continue
			}

			return dataStore, nil
		case <-pollingTimeout.C:
			return nil, errors.New("Get frame ticker exit")
		case err := <-mda.Ctx.Done():
			log.Warn("Retrieval service shutting down", "err", err)
			return nil, errors.New("Retrieval service shutting down")
		}
	}
}
