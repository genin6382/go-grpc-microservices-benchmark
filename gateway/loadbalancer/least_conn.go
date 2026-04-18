package loadbalancer

import (
	"net/http"
	"sync"
)

type LeastConnections struct {
	mu       sync.RWMutex
	backends map[string][]string
	inflight map[string]map[string]int
}

func NewLeastConnections() *LeastConnections {
	return &LeastConnections{
		backends: make(map[string][]string),
		inflight: make(map[string]map[string]int),
	}
}

func (lc *LeastConnections) UpdateBackends(service string, backends []string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	copied := make([]string, len(backends))
	copy(copied, backends)
	lc.backends[service] = copied

	if _, ok := lc.inflight[service]; !ok {
		lc.inflight[service] = make(map[string]int)
	}

	alive := lc.inflight[service]
	for _, b := range copied {
		if _, ok := alive[b]; !ok {
			alive[b] = 0
		}
	}

	for b := range alive {
		found := false
		for _, x := range copied {
			if x == b {
				found = true
				break
			}
		}
		if !found {
			delete(alive, b)
		}
	}
}

func (lc *LeastConnections) NextBackend(service string, r *http.Request) (string, error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	backends := lc.backends[service]
	if len(backends) == 0 {
		return "", ErrNoBackends
	}

	if _, ok := lc.inflight[service]; !ok {
		lc.inflight[service] = make(map[string]int)
	}

	minBackend := backends[0]
	minCount := lc.inflight[service][minBackend]

	for _, b := range backends[1:] {
		if c := lc.inflight[service][b]; c < minCount {
			minBackend = b
			minCount = c
		}
	}

	lc.inflight[service][minBackend]++
	return minBackend, nil
}

func (lc *LeastConnections) Done(service, backend string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if m, ok := lc.inflight[service]; ok {
		if m[backend] > 0 {
			m[backend]--
		}
	}
}

func (lc *LeastConnections) CurrentCount(service, backend string) int {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	if m, ok := lc.inflight[service]; ok {
		return m[backend]
	}
	return 0
}