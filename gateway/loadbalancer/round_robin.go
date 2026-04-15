package loadbalancer

import (
	"net/http"
	"sync"
	"sync/atomic"
)

type RoundRobin struct {
	mu       sync.RWMutex
	backends map[string][]string
	counters map[string]*atomic.Uint64
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{
		backends: make(map[string][]string),
		counters: make(map[string]*atomic.Uint64),
	}
}

func (rr *RoundRobin) UpdateBackends(service string, backends []string) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	copied := make([]string, len(backends))
	copy(copied, backends)
	rr.backends[service] = copied

	if _, exists := rr.counters[service]; !exists {
		rr.counters[service] = &atomic.Uint64{}
	}
}

func (rr *RoundRobin) NextBackend(service string, r *http.Request) (string, error) {
	rr.mu.RLock()
	backends, ok := rr.backends[service]
	counter, counterOK := rr.counters[service]
	rr.mu.RUnlock()

	if !ok || len(backends) == 0 {
		return "", ErrNoBackends
	}

	if !counterOK || counter == nil {
		return "", ErrNoBackends
	}

	idx := counter.Add(1)
	backend := backends[(idx-1)%uint64(len(backends))]
	return backend, nil
}

func (rr *RoundRobin) CurrentCount(service, backend string) int {
	return 0
}

func (rr *RoundRobin) Done(service, backend string) {}