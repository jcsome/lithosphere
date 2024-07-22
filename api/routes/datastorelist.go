package routes

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// DataStoreListHandler ... Handles /api/v1/datastore/list GET requests
func (h Routes) DataStoreListHandler(w http.ResponseWriter, r *http.Request) {
	pageQuery := r.URL.Query().Get("page")
	pageSizeQuery := r.URL.Query().Get("pageSize")
	order := r.URL.Query().Get("order")
	params, err := h.svc.QueryPageListParams(pageQuery, pageSizeQuery, order)
	if err != nil {
		http.Error(w, "invalid query params", http.StatusBadRequest)
		h.logger.Error("error reading request params", "err", err.Error())
		return
	}
	cacheKey := fmt.Sprintf("dataStoreList{page:%s,pageSize:%s,order:%s}", pageQuery, pageSizeQuery, order)
	if h.enableCache {
		response, _ := h.cache.GetDataStoreList(cacheKey)
		if response != nil {
			err = jsonResponse(w, response, http.StatusOK)
			if err != nil {
				h.logger.Error("Error writing response", "err", err.Error())
			}
			return
		}
	}
	dataStoreList, err := h.svc.GetDataStoreList(params)
	if err != nil {
		http.Error(w, "Internal server error reading data store list", http.StatusInternalServerError)
		h.logger.Error("Unable to read data store list from DB", "err", err.Error())
		return
	}
	if h.enableCache {
		h.cache.AddDataStoreList(cacheKey, dataStoreList)
	}
	err = jsonResponse(w, dataStoreList, http.StatusOK)
	if err != nil {
		h.logger.Error("Error writing response", "err", err.Error())
	}
}

// DataStoreByIdHandler ... Handles /api/v1/datastore/id/{id} GET requests
func (h Routes) DataStoreByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	params, err := h.svc.QueryByIdParams(idStr)
	if err != nil {
		http.Error(w, "invalid query params", http.StatusBadRequest)
		h.logger.Error("error reading request params", "err", err.Error())
		return
	}
	cacheKey := fmt.Sprintf("dataStore{id:%s}", idStr)
	if h.enableCache {
		response, _ := h.cache.GetDataStoreById(cacheKey)
		if response != nil {
			err = jsonResponse(w, response, http.StatusOK)
			if err != nil {
				h.logger.Error("Error writing response", "err", err.Error())
			}
			return
		}
	}
	dataStore, err := h.svc.GetDataStoreById(params)
	if err != nil {
		http.Error(w, "Internal server error reading data store by id", http.StatusInternalServerError)
		h.logger.Error("Unable to read data store by id from DB", "err", err.Error())
		return
	}
	if h.enableCache {
		h.cache.AddDataStoreById(cacheKey, dataStore)
	}
	err = jsonResponse(w, dataStore, http.StatusOK)
	if err != nil {
		h.logger.Error("Error writing response", "err", err.Error())
	}
}
