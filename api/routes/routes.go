package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/mantlenetworkio/lithosphere/cache"

	"github.com/ethereum/go-ethereum/log"

	"github.com/mantlenetworkio/lithosphere/api/service"
)

type Routes struct {
	logger      log.Logger
	router      *chi.Mux
	svc         service.Service
	enableCache bool
	cache       *cache.LruCache
}

// NewRoutes ... Construct a new route handler instance
func NewRoutes(l log.Logger, r *chi.Mux, svc service.Service, enableCache bool, cache *cache.LruCache) Routes {
	return Routes{
		logger:      l,
		router:      r,
		svc:         svc,
		enableCache: enableCache,
		cache:       cache,
	}
}
