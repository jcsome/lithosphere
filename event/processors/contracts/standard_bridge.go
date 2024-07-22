package contracts

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/mantlenetworkio/lithosphere/common/bigint"
	"github.com/mantlenetworkio/lithosphere/database"
	"github.com/mantlenetworkio/lithosphere/database/event"
	"github.com/mantlenetworkio/lithosphere/database/utils"
	"github.com/mantlenetworkio/lithosphere/event/op-bindings/bindings"
	"github.com/mantlenetworkio/lithosphere/event/op-bindings/predeploys"
)

type StandardBridgeInitiatedEvent struct {
	Event                  *event.ContractEvent
	CrossDomainMessageHash *common.Hash
	FromAddress            common.Address
	ToAddress              common.Address
	ETHAmount              *big.Int
	ERC20Amount            *big.Int
	Data                   utils.Bytes
	LocalTokenAddress      common.Address
	RemoteTokenAddress     common.Address
	Timestamp              uint64
}

type StandardBridgeFinalizedEvent struct {
	Event                  *event.ContractEvent
	CrossDomainMessageHash *common.Hash
	FromAddress            common.Address
	ToAddress              common.Address
	ETHAmount              *big.Int
	ERC20Amount            *big.Int
	Data                   utils.Bytes
	LocalTokenAddress      common.Address
	RemoteTokenAddress     common.Address
	Timestamp              uint64
}

// StandardBridgeInitiatedEvents extracts all initiated bridge events from the contracts that follow the StandardBridge ABI. The
// correlated CrossDomainMessenger nonce is also parsed from the associated messenger events.
func StandardBridgeInitiatedEvents(chainSelector string, contractAddress common.Address, db *database.DB, fromHeight, toHeight *big.Int) ([]StandardBridgeInitiatedEvent, error) {
	ethBridgeInitiatedEvents, err := _standardBridgeInitiatedEvents[bindings.StandardBridgeETHBridgeInitiated](contractAddress, chainSelector, db, fromHeight, toHeight)
	if err != nil {
		return nil, err
	}

	erc20BridgeInitiatedEvents, err := _standardBridgeInitiatedEvents[bindings.StandardBridgeERC20BridgeInitiated](contractAddress, chainSelector, db, fromHeight, toHeight)
	if err != nil {
		return nil, err
	}

	mntBridgeInitiatedEvents, err := _standardBridgeInitiatedEvents[bindings.StandardBridgeMNTBridgeInitiated](contractAddress, chainSelector, db, fromHeight, toHeight)
	if err != nil {
		return nil, err
	}

	return append(append(ethBridgeInitiatedEvents, erc20BridgeInitiatedEvents...), mntBridgeInitiatedEvents...), nil
}

// StandardBridgeFinalizedEvents extracts all finalization bridge events from the contracts that follow the StandardBridge ABI. The
// correlated CrossDomainMessenger nonce is also parsed by looking at the parameters of the corresponding relayMessage transaction data.
func StandardBridgeFinalizedEvents(chainSelector string, contractAddress common.Address, db *database.DB, fromHeight, toHeight *big.Int) ([]StandardBridgeFinalizedEvent, error) {
	ethBridgeFinalizedEvents, err := _standardBridgeFinalizedEvents[bindings.StandardBridgeETHBridgeFinalized](contractAddress, chainSelector, db, fromHeight, toHeight)
	if err != nil {
		return nil, err
	}

	erc20BridgeFinalizedEvents, err := _standardBridgeFinalizedEvents[bindings.StandardBridgeERC20BridgeFinalized](contractAddress, chainSelector, db, fromHeight, toHeight)
	if err != nil {
		return nil, err
	}

	mntBridgeFinalizedEvents, err := _standardBridgeFinalizedEvents[bindings.StandardBridgeMNTBridgeFinalized](contractAddress, chainSelector, db, fromHeight, toHeight)
	if err != nil {
		return nil, err
	}

	return append(append(ethBridgeFinalizedEvents, erc20BridgeFinalizedEvents...), mntBridgeFinalizedEvents...), nil
}

// parse out eth or erc20 bridge initiated events
func _standardBridgeInitiatedEvents[BridgeEventType bindings.StandardBridgeETHBridgeInitiated | bindings.StandardBridgeERC20BridgeInitiated | bindings.StandardBridgeMNTBridgeInitiated](
	contractAddress common.Address, chainSelector string, db *database.DB, fromHeight, toHeight *big.Int,
) ([]StandardBridgeInitiatedEvent, error) {
	standardBridgeAbi, err := bindings.StandardBridgeMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	var eventType BridgeEventType
	var eventName string
	switch any(eventType).(type) {
	case bindings.StandardBridgeETHBridgeInitiated:
		eventName = "ETHBridgeInitiated"
	case bindings.StandardBridgeERC20BridgeInitiated:
		eventName = "ERC20BridgeInitiated"
	case bindings.StandardBridgeMNTBridgeInitiated:
		eventName = "MNTBridgeInitiated"
	default:
		panic("should not be here")
	}

	initiatedBridgeEventAbi := standardBridgeAbi.Events[eventName]
	contractEventFilter := event.ContractEvent{ContractAddress: contractAddress, EventSignature: initiatedBridgeEventAbi.ID}
	initiatedBridgeEvents, err := db.ContractEvents.ContractEventsWithFilter(contractEventFilter, chainSelector, fromHeight, toHeight)
	if err != nil {
		return nil, err
	}

	standardBridgeInitiatedEvents := make([]StandardBridgeInitiatedEvent, len(initiatedBridgeEvents))
	for i := range initiatedBridgeEvents {
		switch any(eventType).(type) {
		case bindings.StandardBridgeETHBridgeInitiated:
			ethBridge := bindings.StandardBridgeETHBridgeInitiated{Raw: *initiatedBridgeEvents[i].RLPLog}
			err := UnpackLog(&ethBridge, initiatedBridgeEvents[i].RLPLog, eventName, standardBridgeAbi)
			if err != nil {
				return nil, err
			}
			standardBridgeInitiatedEvents[i] = StandardBridgeInitiatedEvent{
				Event:              &initiatedBridgeEvents[i],
				LocalTokenAddress:  predeploys.BVM_ETHAddr,
				RemoteTokenAddress: predeploys.BVM_ETHAddr,
				FromAddress:        ethBridge.From,
				ToAddress:          ethBridge.To,
				ETHAmount:          ethBridge.Amount,
				ERC20Amount:        bigint.Zero,
				Data:               ethBridge.ExtraData,
				Timestamp:          initiatedBridgeEvents[i].Timestamp,
			}
		case bindings.StandardBridgeMNTBridgeInitiated:
			mntBridge := bindings.StandardBridgeMNTBridgeInitiated{Raw: *initiatedBridgeEvents[i].RLPLog}
			err := UnpackLog(&mntBridge, initiatedBridgeEvents[i].RLPLog, eventName, standardBridgeAbi)
			if err != nil {
				return nil, err
			}

			standardBridgeInitiatedEvents[i] = StandardBridgeInitiatedEvent{
				Event:              &initiatedBridgeEvents[i],
				LocalTokenAddress:  predeploys.LegacyERC20MNTAddr,
				RemoteTokenAddress: predeploys.LegacyERC20MNTAddr,
				FromAddress:        mntBridge.From,
				ToAddress:          mntBridge.To,
				ETHAmount:          bigint.Zero,
				ERC20Amount:        mntBridge.Amount,
				Data:               mntBridge.ExtraData,
				Timestamp:          initiatedBridgeEvents[i].Timestamp,
			}
		case bindings.StandardBridgeERC20BridgeInitiated:
			erc20Bridge := bindings.StandardBridgeERC20BridgeInitiated{Raw: *initiatedBridgeEvents[i].RLPLog}
			err := UnpackLog(&erc20Bridge, initiatedBridgeEvents[i].RLPLog, eventName, standardBridgeAbi)
			if err != nil {
				return nil, err
			}
			standardBridgeInitiatedEvents[i] = StandardBridgeInitiatedEvent{
				Event:              &initiatedBridgeEvents[i],
				LocalTokenAddress:  erc20Bridge.LocalToken,
				RemoteTokenAddress: erc20Bridge.RemoteToken,
				FromAddress:        erc20Bridge.From,
				ToAddress:          erc20Bridge.To,
				ETHAmount:          bigint.Zero,
				ERC20Amount:        erc20Bridge.Amount,
				Data:               erc20Bridge.ExtraData,
				Timestamp:          initiatedBridgeEvents[i].Timestamp,
			}
		}
	}
	return standardBridgeInitiatedEvents, nil
}

// parse out eth mnt or erc20 bridge finalization events
func _standardBridgeFinalizedEvents[BridgeEventType bindings.StandardBridgeETHBridgeFinalized | bindings.StandardBridgeERC20BridgeFinalized | bindings.StandardBridgeMNTBridgeFinalized](
	contractAddress common.Address, chainSelector string, db *database.DB, fromHeight, toHeight *big.Int,
) ([]StandardBridgeFinalizedEvent, error) {
	standardBridgeAbi, err := bindings.StandardBridgeMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	var eventType BridgeEventType
	var eventName string
	switch any(eventType).(type) {
	case bindings.StandardBridgeETHBridgeFinalized:
		eventName = "ETHBridgeFinalized"
	case bindings.StandardBridgeERC20BridgeFinalized:
		eventName = "ERC20BridgeFinalized"
	case bindings.StandardBridgeMNTBridgeFinalized:
		eventName = "MNTBridgeFinalized"
	default:
		panic("should not be here")
	}

	bridgeFinalizedEventAbi := standardBridgeAbi.Events[eventName]
	contractEventFilter := event.ContractEvent{ContractAddress: contractAddress, EventSignature: bridgeFinalizedEventAbi.ID}
	bridgeFinalizedEvents, err := db.ContractEvents.ContractEventsWithFilter(contractEventFilter, chainSelector, fromHeight, toHeight)
	if err != nil {
		return nil, err
	}

	standardBridgeFinalizedEvents := make([]StandardBridgeFinalizedEvent, len(bridgeFinalizedEvents))
	for i := range bridgeFinalizedEvents {

		switch any(eventType).(type) {
		case bindings.StandardBridgeETHBridgeFinalized:
			ethBridge := bindings.StandardBridgeETHBridgeFinalized{Raw: *bridgeFinalizedEvents[i].RLPLog}
			err := UnpackLog(&ethBridge, bridgeFinalizedEvents[i].RLPLog, eventName, standardBridgeAbi)
			if err != nil {
				return nil, err
			}

			standardBridgeFinalizedEvents[i] = StandardBridgeFinalizedEvent{
				Event:              &bridgeFinalizedEvents[i],
				LocalTokenAddress:  predeploys.BVM_ETHAddr,
				RemoteTokenAddress: predeploys.BVM_ETHAddr,
				FromAddress:        ethBridge.From,
				ToAddress:          ethBridge.To,
				ETHAmount:          ethBridge.Amount,
				ERC20Amount:        bigint.Zero,
				Data:               ethBridge.ExtraData,
				Timestamp:          bridgeFinalizedEvents[i].Timestamp,
			}

		case bindings.StandardBridgeMNTBridgeFinalized:
			mntBridge := bindings.StandardBridgeMNTBridgeFinalized{Raw: *bridgeFinalizedEvents[i].RLPLog}
			err := UnpackLog(&mntBridge, bridgeFinalizedEvents[i].RLPLog, eventName, standardBridgeAbi)
			if err != nil {
				return nil, err
			}

			standardBridgeFinalizedEvents[i] = StandardBridgeFinalizedEvent{
				Event:              &bridgeFinalizedEvents[i],
				LocalTokenAddress:  predeploys.LegacyERC20MNTAddr,
				RemoteTokenAddress: predeploys.LegacyERC20MNTAddr,
				FromAddress:        mntBridge.From,
				ToAddress:          mntBridge.To,
				ETHAmount:          bigint.Zero,
				ERC20Amount:        mntBridge.Amount,
				Data:               mntBridge.ExtraData,
				Timestamp:          bridgeFinalizedEvents[i].Timestamp,
			}

		case bindings.StandardBridgeERC20BridgeFinalized:
			erc20Bridge := bindings.StandardBridgeERC20BridgeFinalized{Raw: *bridgeFinalizedEvents[i].RLPLog}
			err := UnpackLog(&erc20Bridge, bridgeFinalizedEvents[i].RLPLog, eventName, standardBridgeAbi)
			if err != nil {
				return nil, err
			}

			standardBridgeFinalizedEvents[i] = StandardBridgeFinalizedEvent{
				Event:              &bridgeFinalizedEvents[i],
				LocalTokenAddress:  erc20Bridge.LocalToken,
				RemoteTokenAddress: erc20Bridge.RemoteToken,
				FromAddress:        erc20Bridge.From,
				ToAddress:          erc20Bridge.To,
				ETHAmount:          bigint.Zero,
				ERC20Amount:        erc20Bridge.Amount,
				Data:               erc20Bridge.ExtraData,
				Timestamp:          bridgeFinalizedEvents[i].Timestamp,
			}
		default:
			panic("should not be here")
		}
	}
	return standardBridgeFinalizedEvents, nil
}
