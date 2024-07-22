package exporter

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/ybbus/jsonrpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/mantlenetworkio/lithosphere/config"
	"github.com/mantlenetworkio/lithosphere/database"
	"github.com/mantlenetworkio/lithosphere/database/event"
	"github.com/mantlenetworkio/lithosphere/event/op-bindings/predeploys"
	"github.com/mantlenetworkio/lithosphere/exporter/k8sClient"
	"github.com/mantlenetworkio/lithosphere/exporter/version"
)

var UnknownStatus = "UNKNOWN"

type healthCheck struct {
	mu             *sync.RWMutex
	height         uint64
	healthy        bool
	updateTime     time.Time
	allowedMethods []string
	version        *string
}

type getBlockByNumberResponse struct {
	BaseFeePerGas string `json:"baseFeePerGas"`
	GasUsed       string `json:"gasUsed"`
}

func healthHandler(health *healthCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health.mu.RLock()
		defer health.mu.RUnlock()
		w.Write([]byte(fmt.Sprintf(`{ "healthy": "%t", "version": "%s" }`, health.healthy, *health.version)))
	}
}

type Exporter struct {
	exporterConfig *config.ExporterConfig
	db             *database.DB
	shutdown       context.CancelCauseFunc

	stopped atomic.Bool
}

func NewExporter(exporterConfig config.ExporterConfig, db *database.DB, shutdown context.CancelCauseFunc) (*Exporter, error) {
	return &Exporter{
		exporterConfig: &exporterConfig,
		db:             db,
		shutdown:       shutdown,
	}, nil
}

func (er *Exporter) Start(ctx context.Context) error {
	if er.exporterConfig.Version {
		fmt.Printf("(version=%s, gitcommit=%s)\n", version.Version, version.GitCommit)
		fmt.Printf("(go=%s, date=%s)\n", version.GoVersion, version.BuildDate)
		os.Exit(0)
	}
	log.Infoln("exporter config", er.exporterConfig.ExportAddress, er.exporterConfig.RpcProvider, er.exporterConfig.NetworkLabel)
	log.Infoln("Starting op_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())
	opExporterVersion.WithLabelValues(
		strings.Trim(version.Version, "\""), version.GitCommit, version.GoVersion, version.BuildDate).Inc()
	health := healthCheck{
		mu:             new(sync.RWMutex),
		height:         0,
		healthy:        false,
		updateTime:     time.Now(),
		allowedMethods: nil,
		version:        &UnknownStatus,
	}
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/health", healthHandler(&health))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>OP Exporter</title></head>
		<body>
		<h1>OP Exporter</h1>
		<p><a href="/metrics">Metrics</a></p>
		<p><a href="/health">Health</a></p>
		</body>
		</html>`))
	})
	go er.getBlockNumber(&health)

	if er.exporterConfig.EnableRollUpGasPrices {
		go er.getRollupGasPrices()
	}

	if er.exporterConfig.EnableGasBaseFee {
		go er.getBaseFee()
	}

	if er.exporterConfig.EnableK8sQuery {
		client, err := k8sClient.NewK8sClient()
		if err != nil {
			log.Fatal(err)
			return err
		}
		go er.getSequencerVersion(&health, client)
	}

	go er.metricBridgeAccountChecking()

	go er.metricDepositsAmount()

	go er.metricWithdrawUnclaimedAmount()

	go er.metricWithdrawClaimedAmount()

	go er.MetricTransferBigValueInEthereum()

	go er.MetricTransferBigValueInMantle()

	log.Infoln("Listening on", er.exporterConfig.ExportAddress)
	if err := http.ListenAndServe(er.exporterConfig.ExportAddress, nil); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (er *Exporter) Stop(ctx context.Context) error {
	var result error

	if er.db != nil {
		if err := er.db.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close DB: %w", err))
		}
	}

	er.stopped.Store(true)
	log.Info("exporter service shutdown complete")

	return result
}

func (er *Exporter) Stopped() bool {
	return er.stopped.Load()
}

func (er *Exporter) getSequencerVersion(health *healthCheck, client *kubernetes.Clientset) {
	ns, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Fatalf("Unable to read namespace file: %s", err)
	}
	ticker := time.NewTicker(30 * time.Second)
	for {
		<-ticker.C
		getOpts := metav1.GetOptions{
			TypeMeta:        metav1.TypeMeta{},
			ResourceVersion: "",
		}
		sequencerStatefulSet, err := client.AppsV1().StatefulSets(string(ns)).Get(context.TODO(), "sequencer", getOpts)
		if err != nil {
			health.version = &UnknownStatus
			log.Errorf("Unable to retrieve a sequencer StatefulSet: %s", err)
			continue
		}
		for _, c := range sequencerStatefulSet.Spec.Template.Spec.Containers {
			log.Infof("Checking container %s", c.Name)
			switch {
			case c.Name == "sequencer":
				log.Infof("The sequencer version is: %s", c.Image)
				health.version = &c.Image
			default:
				log.Infof("Unable to find the sequencer container in the statefulset?!?")
			}
		}
	}
}

func (er *Exporter) getBlockNumber(health *healthCheck) {
	rpcClient := jsonrpc.NewClientWithOpts(er.exporterConfig.RpcProvider, &jsonrpc.RPCClientOpts{})
	var blockNumberResponse *string
	for {
		if err := rpcClient.CallFor(&blockNumberResponse, "eth_blockNumber"); err != nil {
			health.mu.Lock()
			health.healthy = false
			health.mu.Unlock()
			log.Warnln("Error calling eth_blockNumber, setting unhealthy", err)
		} else {
			log.Infoln("Got block number: ", *blockNumberResponse)
			health.mu.Lock()
			currentHeight, err := hexutil.DecodeUint64(*blockNumberResponse)
			blockNumber.WithLabelValues(
				er.exporterConfig.NetworkLabel, "layer2").Set(float64(currentHeight))
			if err != nil {
				log.Warnln("Error decoding block height", err)
				continue
			}
			lastHeight := health.height
			// If the currentHeight is the same as the lastHeight, check that
			// the unhealthyTimePeriod has passed and update health.healthy
			if currentHeight == lastHeight {
				currentTime := time.Now()
				lastTime := health.updateTime
				log.Warnln(fmt.Sprintf("Heights are the same, %v, %v", currentTime, lastTime))
				if lastTime.Add(time.Duration(er.exporterConfig.UnhealthyTimePeriod) * time.Minute).Before(currentTime) {
					health.healthy = false
					log.Warnln("Heights are the same for the unhealthyTimePeriod, setting unhealthy")
				}
			} else {
				log.Warnln("New block height detected, setting healthy")
				health.height = currentHeight
				health.updateTime = time.Now()
				health.healthy = true
			}
			if health.healthy {
				healthySequencer.WithLabelValues(
					er.exporterConfig.NetworkLabel).Set(1)
			} else {
				healthySequencer.WithLabelValues(
					er.exporterConfig.NetworkLabel).Set(0)
			}

			health.mu.Unlock()
		}
		time.Sleep(time.Duration(er.exporterConfig.SequencerPollingSeconds) * time.Second)
	}
}

func (er *Exporter) getBaseFee() {
	rpcClient := jsonrpc.NewClientWithOpts(er.exporterConfig.RpcProvider, &jsonrpc.RPCClientOpts{})
	var getBlockByNumbeResponse *getBlockByNumberResponse
	for {
		if err := rpcClient.CallFor(&getBlockByNumbeResponse, "eth_getBlockByNumber", "latest", false); err != nil {
			log.Errorln("Error calling eth_getBlockByNumber", err)
		} else {
			log.Infoln("Got baseFee response: ", *getBlockByNumbeResponse)
			baseFeePerGas, err := hexutil.DecodeUint64(getBlockByNumbeResponse.BaseFeePerGas)
			if err != nil {
				log.Warnln("Error converting baseFeePerGas " + getBlockByNumbeResponse.BaseFeePerGas)
			}
			gasBaseFeeMetric.WithLabelValues(
				er.exporterConfig.NetworkLabel, "layer2").Set(float64(baseFeePerGas))

			gasUsed, err := hexutil.DecodeUint64(getBlockByNumbeResponse.GasUsed)
			if err != nil {
				log.Warnln("Error converting gasUsed " + getBlockByNumbeResponse.GasUsed)
			}
			gasUsedMetric.WithLabelValues(
				er.exporterConfig.NetworkLabel, "layer2").Set(float64(gasUsed))
		}
		time.Sleep(time.Duration(er.exporterConfig.SequencerPollingSeconds) * time.Second)

	}
}

func (er *Exporter) getRollupGasPrices() {
	rpcClient := jsonrpc.NewClientWithOpts(er.exporterConfig.RpcProvider, &jsonrpc.RPCClientOpts{})
	var rollupGasPrices *GetRollupGasPrices
	for {
		if err := rpcClient.CallFor(&rollupGasPrices, "rollup_gasPrices"); err != nil {
			log.Warnln("Error calling rollup_gasPrices", err)
		} else {
			l1GasPriceString := rollupGasPrices.L1GasPrice
			l1GasPrice, err := hexutil.DecodeUint64(l1GasPriceString)
			if err != nil {
				log.Warnln("Error converting gasPrice " + l1GasPriceString)
			}
			gasPrice.WithLabelValues(
				er.exporterConfig.NetworkLabel, "layer1").Set(float64(l1GasPrice))
			l2GasPriceString := rollupGasPrices.L2GasPrice
			l2GasPrice, err := hexutil.DecodeUint64(l2GasPriceString)
			if err != nil {
				log.Warnln("Error converting gasPrice " + l2GasPriceString)
			}
			gasPrice.WithLabelValues(
				er.exporterConfig.NetworkLabel, "layer2").Set(float64(l2GasPrice))
			log.Infoln("Got L1 gas string: ", l1GasPriceString)
			log.Infoln("Got L1 gas prices: ", l1GasPrice)
			log.Infoln("Got L2 gas string: ", l2GasPriceString)
			log.Infoln("Got L2 gas prices: ", l2GasPrice)
		}
		time.Sleep(time.Duration(30) * time.Second)
	}
}

func (er *Exporter) metricBridgeAccountChecking() {
	ticker := time.NewTicker(15 * time.Minute)
	for {
		<-ticker.C
		checkpoints := er.db.CheckPoint.GetLatestBridgeCheckpoint()
		for _, checkpoint := range checkpoints {
			symbol, _ := er.db.TokenList.GetSymbolByAddress(checkpoint.L1TokenAddress)
			bridgeAccountChecking.WithLabelValues(checkpoint.L1TokenAddress, checkpoint.L2TokenAddress, symbol).Set(float64(checkpoint.Status))
		}
	}
}

func (er *Exporter) metricDepositsAmount() {
	ticker := time.NewTicker(30 * time.Second)
	startTimestamp := 0
	endTimestamp := 0
	for {
		<-ticker.C
		endTimestamp = er.db.L1ToL2.L1L2LatestTimestamp()
		deposits, err := er.db.L1ToL2.GetDepositsAmountByTimestamp(startTimestamp, endTimestamp)
		if err != nil {
			log.Errorln(err.Error())
			continue
		}

		for _, deposit := range deposits {
			log.Infoln("get deposit amount", deposit.L1TokenAddress, deposit.L2TokenAddress)
			symbol, _ := er.db.TokenList.GetSymbolByAddress(deposit.L1TokenAddress.String())
			if deposit.L1TokenAddress.String() == predeploys.BVM_ETH {
				depositsAmount.WithLabelValues(strings.ToLower(deposit.L1TokenAddress.String()), symbol).Add(float64(deposit.ETHAmount.Uint64()))
			} else {
				depositsAmount.WithLabelValues(strings.ToLower(deposit.L1TokenAddress.String()), symbol).Add(float64(deposit.ERC20Amount.Uint64()))
			}
		}
		startTimestamp = endTimestamp
	}
}

func (er *Exporter) metricWithdrawUnclaimedAmount() {
	ticker := time.NewTicker(30 * time.Second)
	for {
		<-ticker.C
		withdraws, err := er.db.L2ToL1.GetWithdrawsUnclaimedAmount(common.Hash{}.String())
		if err != nil {
			log.Errorln(err)
			continue
		}
		if len(withdraws) == 0 {
			WithdrawBigValueAddresses := strings.Split(er.exporterConfig.WithdrawBigValueAddress, " ")
			for _, add := range WithdrawBigValueAddresses {
				symbol, _ := er.db.TokenList.GetSymbolByAddress(add)
				floatVal := float64(0)
				withdrawsUnclaimedAmount.WithLabelValues(add, symbol).Set(floatVal)
			}
			log.Infoln("no more withdraw unclaimed transactions")
		}

		for _, withdraw := range withdraws {
			log.Infoln("get withdraw unclaimed amount", withdraw.L1TokenAddress, withdraw.L2TokenAddress)
			symbol, _ := er.db.TokenList.GetSymbolByAddress(withdraw.L1TokenAddress.String())
			if withdraw.L1TokenAddress.String() == predeploys.BVM_ETH {
				withdrawsUnclaimedAmount.WithLabelValues(strings.ToLower(withdraw.L1TokenAddress.String()), symbol).Set(float64(withdraw.ETHAmount.Uint64()))
			} else {
				withdrawsUnclaimedAmount.WithLabelValues(strings.ToLower(withdraw.L1TokenAddress.String()), symbol).Set(float64(withdraw.ERC20Amount.Uint64()))
			}
		}
	}
}

func (er *Exporter) metricWithdrawClaimedAmount() {
	ticker := time.NewTicker(30 * time.Second)
	startBlockNumber := 0
	endBlockNumber := 0
	for {
		<-ticker.C
		endBlockNumber = er.db.L2ToL1.L2L1LatestFinalizedL1BlockNumber()
		withdraws, err := er.db.L2ToL1.GetWithdrawsClaimedAmount(common.Hash{}.String(), startBlockNumber, endBlockNumber)
		if err != nil {
			log.Errorln(err.Error())
			continue
		}

		if len(withdraws) == 0 {
			WithdrawBigValueAddresses := strings.Split(er.exporterConfig.WithdrawBigValueAddress, " ")
			for _, add := range WithdrawBigValueAddresses {
				symbol, _ := er.db.TokenList.GetSymbolByAddress(add)
				floatVal, _ := strconv.ParseFloat("0", 64)
				withdrawsClaimedAmount.WithLabelValues(add, symbol).Add(floatVal)
			}
			log.Infoln("no more withdraw claimed transactions")
		}
		for _, withdraw := range withdraws {
			log.Infoln("get withdraw claimed amount", withdraw.L1TokenAddress, withdraw.L2TokenAddress)
			symbol, _ := er.db.TokenList.GetSymbolByAddress(withdraw.L1TokenAddress.String())
			if withdraw.L1TokenAddress.String() == predeploys.BVM_ETH {
				withdrawsClaimedAmount.WithLabelValues(strings.ToLower(withdraw.L1TokenAddress.String()), symbol).Add(float64(withdraw.ETHAmount.Uint64()))
			} else {
				withdrawsClaimedAmount.WithLabelValues(strings.ToLower(withdraw.L1TokenAddress.String()), symbol).Add(float64(withdraw.ERC20Amount.Uint64()))
			}
		}
		startBlockNumber = endBlockNumber
	}
}

func (er *Exporter) MetricTransferBigValueInEthereum() {
	ticker := time.NewTicker(30 * time.Second)
	for {
		<-ticker.C
		transferBigValueAddresses := strings.Split(er.exporterConfig.TransferBigValueAddressInEthereum, " ")
		transferBigValues := strings.Split(er.exporterConfig.TransferBigValueInEthereum, " ")
		if transferBigValueAddresses[0] == "" {
			return
		}

		contractAddrTransferMap := make(map[string]map[string]*big.Int)

		tsNow := time.Now().Unix()
		tsBefore := time.Now().Add(-time.Minute * 5).Unix()

		for _, addr := range transferBigValueAddresses {

			l1Events, err := er.db.ContractEvents.L1ContractEventsWithContractFilter(addr, event.TransferEventABIHash, tsBefore, tsNow)
			if err != nil {
				log.Errorln(err)
			}
			addressAllTransferMap := make(map[string]*big.Int)
			for _, event := range l1Events {
				topic1 := hex.EncodeToString(event.RLPLog.Topics[1].Bytes())
				topic2 := hex.EncodeToString(event.RLPLog.Topics[2].Bytes())

				if strings.Compare(topic1, "0000000000000000000000000000000000000000000000000000000000000000") == 0 || strings.Compare(topic2, "0000000000000000000000000000000000000000000000000000000000000000") == 0 {
					continue
				}
				num := new(big.Int)
				num.SetString(hex.EncodeToString(event.RLPLog.Data[2:]), 16)
				if existingData, ok := addressAllTransferMap[topic1]; ok {
					addressAllTransferMap[topic1] = new(big.Int).Add(existingData, num)
				} else {
					addressAllTransferMap[topic1] = num
				}
			}
			contractAddrTransferMap[addr] = addressAllTransferMap
		}

		for i, addr := range transferBigValueAddresses {
			var alertNum = 0
			symbol, _ := er.db.TokenList.GetSymbolByAddress(addr)

			transferBigValue := transferBigValues[i]
			for userAddress, value := range contractAddrTransferMap[addr] {

				bigValue := new(big.Int)
				bigValue.SetString(transferBigValue, 10)
				if value.Cmp(bigValue) != -1 {
					alertNum++
					log.Info(symbol + "出现大额交易：" + userAddress + "在***时间内转移" + value.String() + symbol)
				}
			}

			transferBigValueInEthereum.WithLabelValues(addr, symbol).Set(float64(alertNum))
		}
	}
}

func (er *Exporter) MetricTransferBigValueInMantle() {

	ticker := time.NewTicker(30 * time.Second)
	for {
		<-ticker.C
		transferBigValueAddresses := strings.Split(er.exporterConfig.TransferBigValueAddressInMantle, " ")
		transferBigValues := strings.Split(er.exporterConfig.TransferBigValueInMantle, " ")
		if transferBigValueAddresses[0] == "" {
			return
		}

		contractAddrTransferMap := make(map[string]map[string]*big.Int)

		tsNow := time.Now().Unix()
		interval := time.Minute * 5
		tsBefore := time.Now().Add(-interval).Unix()

		for _, addr := range transferBigValueAddresses {

			l2Events, err := er.db.ContractEvents.L2ContractEventsWithContractFilter(addr, event.TransferEventABIHash, tsBefore, tsNow)
			if err != nil {
				log.Errorln(err)
			}
			addressAllTransferMap := make(map[string]*big.Int)
			for _, event := range l2Events {
				topic1 := hex.EncodeToString(event.RLPLog.Topics[1].Bytes())
				topic2 := hex.EncodeToString(event.RLPLog.Topics[2].Bytes())

				if strings.Compare(topic1, "0000000000000000000000000000000000000000000000000000000000000000") == 0 || strings.Compare(topic2, "0000000000000000000000000000000000000000000000000000000000000000") == 0 {
					continue
				}
				num := new(big.Int)
				num.SetString(hex.EncodeToString(event.RLPLog.Data[2:]), 16)
				if existingData, ok := addressAllTransferMap[topic1]; ok {
					addressAllTransferMap[topic1] = new(big.Int).Add(existingData, num)
				} else {
					addressAllTransferMap[topic1] = num
				}
			}
			contractAddrTransferMap[addr] = addressAllTransferMap
		}

		for i, addr := range transferBigValueAddresses {
			var alertNum = 0

			symbol, _ := er.db.TokenList.GetSymbolByAddress(addr)
			transferBigValue := transferBigValues[i]
			for userAddress, value := range contractAddrTransferMap[addr] {

				bigValue := new(big.Int)
				bigValue.SetString(transferBigValue, 10)
				if value.Cmp(bigValue) != -1 {
					alertNum++
					log.Infof("%s出现大额交易：%s在%s时间内转移%s", symbol, userAddress, interval, value)
				}
			}

			transferBigValueInMantle.WithLabelValues(addr, symbol).Set(float64(alertNum))
		}
	}
}
