package routes

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// StateRootListHandler ... Handles "/api/v1/stateroot/list" GET requests
func (h Routes) StateRootListHandler(w http.ResponseWriter, r *http.Request) {
	pageQuery := r.URL.Query().Get("page")
	pageSizeQuery := r.URL.Query().Get("pageSize")
	order := r.URL.Query().Get("order")
	params, err := h.svc.QueryPageListParams(pageQuery, pageSizeQuery, order)
	if err != nil {
		http.Error(w, "invalid query params", http.StatusBadRequest)
		h.logger.Error("error reading request params", "err", err.Error())
		return
	}

	cacheKey := fmt.Sprintf("stateRootList{page:%s,pageSize:%s,order:%s}", pageQuery, pageSizeQuery, order)
	if h.enableCache {
		response, _ := h.cache.GetStateRootList(cacheKey)
		if response != nil {
			err = jsonResponse(w, response, http.StatusOK)
			if err != nil {
				h.logger.Error("Error writing response", "err", err.Error())
			}
			return
		}
	}
	stateRootPage, err := h.svc.GetStateRootList(params)
	if err != nil {
		http.Error(w, "Internal server error reading state root list", http.StatusInternalServerError)
		h.logger.Error("Unable to read state root list from DB", "err", err.Error())
		return
	}
	if h.enableCache {
		h.cache.AddStateRootList(cacheKey, stateRootPage)
	}
	err = jsonResponse(w, stateRootPage, http.StatusOK)
	if err != nil {
		h.logger.Error("Error writing response", "err", err.Error())
	}
}

// StateRootByIndexHandler ... Handles /api/v1/stateroot/index/{index} GET requests
func (h Routes) StateRootByIndexHandler(w http.ResponseWriter, r *http.Request) {
	indexStr := chi.URLParam(r, "index")

	params, err := h.svc.QueryByIndexParams(indexStr)
	if err != nil {
		http.Error(w, "invalid query params", http.StatusBadRequest)
		h.logger.Error("error reading request params", "err", err.Error())
		return
	}

	cacheKey := fmt.Sprintf("stateRootByIndex{index:%s}", indexStr)
	if h.enableCache {
		response, _ := h.cache.GetStateRootByIndex(cacheKey)
		if response != nil {
			err = jsonResponse(w, response, http.StatusOK)
			if err != nil {
				h.logger.Error("Error writing response", "err", err.Error())
			}
			return
		}
	}

	stateRoot, err := h.svc.GetStateRootByIndex(params)
	if err != nil {
		http.Error(w, "Internal server error reading state root list", http.StatusInternalServerError)
		h.logger.Error("Unable to read state root list from DB", "err", err.Error())
		return
	}
	if h.enableCache {
		h.cache.AddStateRootByIndex(cacheKey, stateRoot)
	}

	err = jsonResponse(w, stateRoot, http.StatusOK)
	if err != nil {
		h.logger.Error("Error writing response", "err", err.Error())
	}
}
