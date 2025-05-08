package roundrobin

import (
	"log/slog"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"testtask/balancer/core"
	"time"
)

type RoundRobin struct {
	log     *slog.Logger
	mu      sync.RWMutex
	servers []core.Server
	current uint64
}

func NewRoundRobin(log *slog.Logger, servers []core.Server) *RoundRobin {
	return &RoundRobin{
		log:     log,
		servers: servers,
	}
}

// Добавляем сервер в пул серверов
func (r *RoundRobin) AddServer(server core.Server) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.servers = append(r.servers, server)
}

// Атомарно возвращаем индекс следующего сервера и инкрементируем счетчик
func (r *RoundRobin) GetNextIndex() int {
	return int(atomic.AddUint64(&r.current, 1))
}

func (r *RoundRobin) ChangeServerStatus(serverUrl *url.URL, status bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, server := range r.servers {
		if server.GetUrl().String() == serverUrl.String() {
			server.SetStatus(status)
			break
		}
	}
}

func (r *RoundRobin) GetNextServer() core.Server {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.servers) == 0 {
		return nil
	}

	// Проходим полный "круг" в происках рабочего сервера
	idxStart := r.GetNextIndex()
	idxEnd := len(r.servers) + idxStart

	// В цикле ищем первый сервер, который в рабочем состоянии
	for i := idxStart; i < idxEnd; i++ {
		idx := i % len(r.servers)
		if r.servers[idx].IsWorking() {
			if i != idxStart {
				atomic.StoreUint64(&r.current, uint64(idx))
			}
			return r.servers[idx]
		}
	}
	return nil
}

// Метод для проверки состояниий серверов в пуле
func (r *RoundRobin) HealthCheck() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, server := range r.servers {
		stat := "working"
		status := r.IsServerWorking(server.GetUrl())
		server.SetStatus(status)
		if !status {
			stat = "failed"
		}
		r.log.Info("Server status:", server.GetUrl().String(), stat)
	}
}

// Метод для установления соединения с конкретным сервером, чтобы проверить его состояние
func (r *RoundRobin) IsServerWorking(url *url.URL) bool {
	timeout := 3 * time.Second
	conn, err := net.DialTimeout("tcp", url.Host, timeout)
	if err != nil {
		r.log.Error("Server failed", "url", url, "error", err)
		return false
	}
	_ = conn.Close()
	return true
}
