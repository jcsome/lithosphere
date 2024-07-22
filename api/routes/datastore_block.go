package routes

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// DataStoreBlockByIDHandler ... Handles /api/v1/datastore/transaction/id/{id} GET requests
func (h Routes) DataStoreBlockByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	params, err := h.svc.QueryByIdParams(idStr)
	if err != nil {
		http.Error(w, "invalid query params", http.StatusBadRequest)
		h.logger.Error("error reading request params", "err", err.Error())
		return
	}

	cacheKey := fmt.Sprintf("dataStoreBlock{id:%s}", idStr)
	if h.enableCache {
		response, _ := h.cache.GetDataStoreBlockById(cacheKey)
		if response != nil {
			err = jsonResponse(w, response, http.StatusOK)
			if err != nil {
				h.logger.Error("Error writing response", "err", err.Error())
			}
			return
		}
	}

	dataStoreBlock, err := h.svc.GetDataStoreBlockByDataStoreId(params)
	if err != nil {
		http.Error(w, "Internal server error reading data store block by id", http.StatusInternalServerError)
		h.logger.Error("Unable to read data store block by id from DB", "err", err.Error())
		return
	}
	if h.enableCache {
		h.cache.AddDataStoreBlockById(cacheKey, dataStoreBlock)
	}

	err = jsonResponse(w, dataStoreBlock, http.StatusOK)
	if err != nil {
		h.logger.Error("Error writing response", "err", err.Error())
	}
}
