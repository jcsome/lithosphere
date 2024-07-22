package cache

import (
	"errors"

	"github.com/acmestack/gorm-plus/gplus"
	lru "github.com/hashicorp/golang-lru/v2/expirable"

	"github.com/mantlenetworkio/lithosphere/api/models"
	"github.com/mantlenetworkio/lithosphere/config"
	"github.com/mantlenetworkio/lithosphere/database/business"
)

/*
apiRouter.Get(fmt.Sprintf(DataStoreListPath), h.DataStoreListHandler)
apiRouter.Get(fmt.Sprintf(DepositsV1Path), h.L1ToL2ListHandler)
apiRouter.Get(fmt.Sprintf(WithdrawalsV1Path), h.L2ToL1ListHandler)
apiRouter.Get(fmt.Sprintf(DataStoreByIDPath+idParam), h.DataStoreByIdHandler)
apiRouter.Get(fmt.Sprintf(DataStoreTxByIDPath+idParam), h.DataStoreBlockByIDHandler)
apiRouter.Get(fmt.Sprintf(StateRootListPath), h.StateRootListHandler)
apiRouter.Get(fmt.Sprintf(StateRootByIndexPath+indexParam), h.StateRootByIndexHandler)
*/

type LruCache struct {
	lruDataStoreList      *lru.LRU[string, any]
	lruL1ToL2List         *lru.LRU[string, any]
	lruL2ToL1List         *lru.LRU[string, any]
	lruStateRootList      *lru.LRU[string, any]
	lruDataStoreById      *lru.LRU[string, any]
	lruDataStoreBlockById *lru.LRU[string, any]
	lruStateRootByIndex   *lru.LRU[string, any]
}

func NewLruCache(cfg config.CacheConfig) *LruCache {
	lruDataStoreList := lru.NewLRU[string, any](cfg.ListSize, nil, cfg.ListExpireTime)
	lruL1ToL2List := lru.NewLRU[string, any](cfg.ListSize, nil, cfg.ListExpireTime)
	lruL2ToL1List := lru.NewLRU[string, any](cfg.ListSize, nil, cfg.ListExpireTime)
	lruStateRootList := lru.NewLRU[string, any](cfg.ListSize, nil, cfg.ListExpireTime)
	lruDataStoreById := lru.NewLRU[string, any](cfg.DetailSize, nil, cfg.DetailExpireTime)
	lruDataStoreBlockById := lru.NewLRU[string, any](cfg.DetailSize, nil, cfg.DetailExpireTime)
	lruStateRootByIndex := lru.NewLRU[string, any](cfg.DetailSize, nil, cfg.DetailExpireTime)
	return &LruCache{
		lruDataStoreList:      lruDataStoreList,
		lruL1ToL2List:         lruL1ToL2List,
		lruL2ToL1List:         lruL2ToL1List,
		lruStateRootList:      lruStateRootList,
		lruDataStoreById:      lruDataStoreById,
		lruDataStoreBlockById: lruDataStoreBlockById,
		lruStateRootByIndex:   lruStateRootByIndex,
	}
}

func (lc *LruCache) GetDataStoreList(key string) (*models.DataStoresResponse, error) {
	result, ok := lc.lruDataStoreList.Get(key)
	if !ok {
		return nil, errors.New("lru get store list fail")
	}
	return result.(*models.DataStoresResponse), nil
}

func (lc *LruCache) AddDataStoreList(key string, data *models.DataStoresResponse) {
	lc.lruDataStoreList.Add(key, data)
}

func (lc *LruCache) GetL1ToL2List(key string) (*models.DepositsResponse, error) {
	result, ok := lc.lruL1ToL2List.Get(key)
	if !ok {
		return nil, errors.New("lru get L1ToL2 list fail")
	}
	return result.(*models.DepositsResponse), nil
}

func (lc *LruCache) AddL1ToL2List(key string, data *models.DepositsResponse) {
	lc.lruL1ToL2List.Add(key, data)
}

func (lc *LruCache) GetL2ToL1List(key string) (*models.WithdrawsResponse, error) {
	result, ok := lc.lruL2ToL1List.Get(key)
	if !ok {
		return nil, errors.New("lru get L2ToL1 list fail")
	}
	return result.(*models.WithdrawsResponse), nil
}

func (lc *LruCache) AddL2ToL1List(key string, data *models.WithdrawsResponse) {
	lc.lruL2ToL1List.Add(key, data)
}

func (lc *LruCache) GetStateRootList(key string) (*gplus.Page[business.StateRoot], error) {
	result, ok := lc.lruStateRootList.Get(key)
	if !ok {
		return nil, errors.New("lru get state root list fail")
	}
	return result.(*gplus.Page[business.StateRoot]), nil
}

func (lc *LruCache) AddStateRootList(key string, data *models.StateRootListResponse) {
	lc.lruStateRootList.Add(key, data)
}

func (lc *LruCache) GetDataStoreById(key string) (*business.DataStore, error) {
	result, ok := lc.lruDataStoreById.Get(key)
	if !ok {
		return nil, errors.New("lru get data store by id fail")
	}
	return result.(*business.DataStore), nil
}

func (lc *LruCache) AddDataStoreById(key string, data *business.DataStore) {
	lc.lruDataStoreById.Add(key, data)
}

func (lc *LruCache) GetDataStoreBlockById(key string) ([]business.DataStoreBlock, error) {
	result, ok := lc.lruDataStoreBlockById.Get(key)
	if !ok {
		return nil, errors.New("lru get data store block by id fail")
	}
	return result.([]business.DataStoreBlock), nil
}

func (lc *LruCache) AddDataStoreBlockById(key string, data []business.DataStoreBlock) {
	lc.lruDataStoreBlockById.Add(key, data)
}

func (lc *LruCache) GetStateRootByIndex(key string) (*business.StateRoot, error) {
	result, ok := lc.lruStateRootByIndex.Get(key)
	if !ok {
		return nil, errors.New("lru get state root by index fail")
	}
	return result.(*business.StateRoot), nil
}

func (lc *LruCache) AddStateRootByIndex(key string, data *business.StateRoot) {
	lc.lruStateRootByIndex.Add(key, data)
}
