package core

import (
	"log/slog"
	"net/http"
	"time"
)

type LoadBalancer struct {
	log        *slog.Logger
	serverPool Pooler
}

func NewLoadBalancer(log *slog.Logger, pool Pooler) *LoadBalancer {
	return &LoadBalancer{
		serverPool: pool,
		log:        log,
	}
}

func (lb *LoadBalancer) Handler(w http.ResponseWriter, r *http.Request) {
	server := lb.serverPool.GetNextServer()
	if server != nil {
		server.GetReverseProxy().ServeHTTP(w, r)
		return
	}
	http.Error(w, "Server not available", http.StatusServiceUnavailable)
}

func (lb *LoadBalancer) StartHealthCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			lb.log.Info("Starting health check")
			lb.serverPool.HealthCheck()
			lb.log.Info("Health check completed")
		}
	}()
}

func (lb *LoadBalancer) Initialize(servers []Server) error {
	if len(servers) == 0 {
		return ErrNoBackends
	}

	for _, server := range servers {
		if server.GetUrl() == nil {
			lb.log.Error("Server has nil URL")
			continue
		}

		lb.serverPool.AddServer(server)
		lb.log.Info("Added server to pool", "url", server.GetUrl().String())
	}

	return nil
}
