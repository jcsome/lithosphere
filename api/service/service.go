package service

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/log"
	"github.com/mantlenetworkio/lithosphere/api/models"
	"github.com/mantlenetworkio/lithosphere/database/business"
	"github.com/mantlenetworkio/lithosphere/database/common"
)

type Service interface {
	GetDepositList(*models.QueryDWParams) (*models.DepositsResponse, error)
	GetWithdrawalList(params *models.QueryDWParams) (*models.WithdrawsResponse, error)
	GetDataStoreList(*models.QueryPageParams) (*models.DataStoresResponse, error)
	GetDataStoreById(params *models.QueryIdParams) (*business.DataStore, error)
	GetDataStoreBlockByDataStoreId(params *models.QueryIdParams) ([]business.DataStoreBlock, error)
	GetStateRootList(*models.QueryPageParams) (*models.StateRootListResponse, error)
	GetStateRootByIndex(*models.QueryIndexParams) (*business.StateRoot, error)

	QueryDWListParams(address string, page string, pageSize string, order string) (*models.QueryDWParams, error)
	QueryPageListParams(page string, pageSize string, order string) (*models.QueryPageParams, error)
	QueryByIdParams(id string) (*models.QueryIdParams, error)
	QueryByIndexParams(index string) (*models.QueryIndexParams, error)
}

type HandlerSvc struct {
	logger        log.Logger
	v             *Validator
	dataStoreView business.DataStoreView
	l1ToL2View    business.L1ToL2View
	l2ToL1View    business.L2ToL1View
	stateRootView business.StateRootView
	blocksView    common.BlocksView
}

func New(v *Validator, dsv business.DataStoreView, l1l2v business.L1ToL2View, l2l1v business.L2ToL1View, blv common.BlocksView, srv business.StateRootView, l log.Logger) Service {
	return &HandlerSvc{
		logger:        l,
		v:             v,
		dataStoreView: dsv,
		l1ToL2View:    l1l2v,
		l2ToL1View:    l2l1v,
		stateRootView: srv,
		blocksView:    blv,
	}
}

func (h HandlerSvc) GetDepositList(params *models.QueryDWParams) (*models.DepositsResponse, error) {
	addressToLower := strings.ToLower(params.Address)
	l1L2List, total := h.l1ToL2View.L1ToL2List(addressToLower, params.Page, params.PageSize, params.Order)
	return &models.DepositsResponse{
		Current: params.Page,
		Size:    params.PageSize,
		Total:   total,
		Records: l1L2List,
	}, nil
}

func (h HandlerSvc) GetWithdrawalList(params *models.QueryDWParams) (*models.WithdrawsResponse, error) {
	addressToLower := strings.ToLower(params.Address)
	l2L1List, total := h.l2ToL1View.L2ToL1List(addressToLower, params.Page, params.PageSize, params.Order)
	return &models.WithdrawsResponse{
		Current: params.Page,
		Size:    params.PageSize,
		Total:   total,
		Records: l2L1List,
	}, nil
}

func (h HandlerSvc) GetDataStoreList(params *models.QueryPageParams) (*models.DataStoresResponse, error) {
	dsList, total := h.dataStoreView.DataStoreList(params.Page, params.PageSize, params.Order)
	items := make([]models.DataStoreList, len(dsList))
	for i, dataStore := range dsList {
		item := models.DataStoreList{
			ID:        dataStore.DataStoreId,
			DataSize:  dataStore.DataSize.Uint64(),
			Status:    dataStore.ConfirmGasUsed.Uint64() != 0,
			Timestamp: dataStore.Timestamp,
			DaHash:    dataStore.DataHash,
		}
		items[i] = item
	}
	return &models.DataStoresResponse{
		Current: params.Page,
		Size:    params.PageSize,
		Total:   total,
		Records: items,
	}, nil
}

func (h HandlerSvc) GetDataStoreById(params *models.QueryIdParams) (*business.DataStore, error) {
	return h.dataStoreView.DataStoreById(big.NewInt(int64(params.Id)))
}

func (h HandlerSvc) GetDataStoreBlockByDataStoreId(params *models.QueryIdParams) ([]business.DataStoreBlock, error) {
	return h.dataStoreView.DataStoreBlockById(big.NewInt(int64(params.Id)))
}

func (h HandlerSvc) GetStateRootList(params *models.QueryPageParams) (*models.StateRootListResponse, error) {
	stateRootList, total := h.stateRootView.StateRootList(params.Page, params.PageSize, params.Order)
	return &models.StateRootListResponse{
		Current: params.Page,
		Size:    params.PageSize,
		Total:   total,
		Records: stateRootList,
	}, nil
}

func (h HandlerSvc) GetStateRootByIndex(params *models.QueryIndexParams) (*business.StateRoot, error) {
	return h.stateRootView.StateRootByIndex(big.NewInt(int64(params.Index)))
}

func (h HandlerSvc) QueryDWListParams(address string, page string, pageSize string, order string) (*models.QueryDWParams, error) {
	var paraAddress string
	if address == "0x00" {
		paraAddress = "0x00"
	} else {
		addr, err := h.v.ParseValidateAddress(address)
		if err != nil {
			h.logger.Error("invalid address param", "address", address, "err", err)
			return nil, err
		}
		paraAddress = addr.String()
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return nil, err
	}
	pageVal := h.v.ValidatePage(pageInt)

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		return nil, err
	}
	pageSizeVal := h.v.ValidatePageSize(pageSizeInt)
	orderBy := h.v.ValidateOrder(order)

	return &models.QueryDWParams{
		Address:  paraAddress,
		Page:     pageVal,
		PageSize: pageSizeVal,
		Order:    orderBy,
	}, nil
}

func (h HandlerSvc) QueryPageListParams(page string, pageSize string, order string) (*models.QueryPageParams, error) {
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return nil, err
	}
	pageValue := h.v.ValidatePage(pageInt)
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		return nil, err
	}
	pageSizeValue := h.v.ValidatePageSize(pageSizeInt)
	orderBy := h.v.ValidateOrder(order)
	return &models.QueryPageParams{
		Page:     pageValue,
		PageSize: pageSizeValue,
		Order:    orderBy,
	}, nil
}

func (h HandlerSvc) QueryByIdParams(id string) (*models.QueryIdParams, error) {
	idValue, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, errors.New("id must be an integer value")
	}
	err = h.v.ValidateIdOrIndex(idValue)
	if err != nil {
		h.logger.Error("invalid query param", "id", id, "err", err)
		return nil, err
	}
	return &models.QueryIdParams{
		Id: idValue,
	}, nil
}

func (h HandlerSvc) QueryByIndexParams(index string) (*models.QueryIndexParams, error) {
	indexValue, err := strconv.ParseUint(index, 10, 64)
	if err != nil {
		return nil, errors.New("index must be an integer value")
	}
	err = h.v.ValidateIdOrIndex(indexValue)
	if err != nil {
		h.logger.Error("invalid query param", "index", index, "err", err)
		return nil, err
	}
	return &models.QueryIndexParams{
		Index: indexValue,
	}, nil
}
