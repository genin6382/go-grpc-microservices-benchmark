package loadbalancer

import (
	"crypto/sha1"
	"encoding/binary"
	"hash/fnv"
	"net/http"
	"sort"
	"strconv"
	"sync"
	log "github.com/sirupsen/logrus"

	internalmiddleware "github.com/genin6382/go-grpc-microservices-benchmark/internal/middleware"
)

type ConsistentHash struct {
	mu            sync.RWMutex
	virtualNodes  int
	backends      map[string][]string
	rings         map[string][]uint32
	ringToBackend map[string]map[uint32]string
}

func NewConsistentHash(virtualNodes int) *ConsistentHash {
	return &ConsistentHash{
		virtualNodes:  virtualNodes,
		backends:      make(map[string][]string),
		rings:         make(map[string][]uint32),
		ringToBackend: make(map[string]map[uint32]string),
	}
}

func (ch *ConsistentHash) UpdateBackends(service string, backends []string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	copied := make([]string, len(backends))
	copy(copied, backends)
	ch.backends[service] = copied

	var ring []uint32
	nodeMap := make(map[uint32]string)

	for _, backend := range copied {
		for i := 0; i < ch.virtualNodes; i++ {
			key := backend + "#" + strconv.Itoa(i)
			h := fnv.New32a()
			_, _ = h.Write([]byte(key))
			hash := h.Sum32()
			ring = append(ring, hash)
			nodeMap[hash] = backend
		}
	}

	sort.Slice(ring, func(i, j int) bool { return ring[i] < ring[j] })
	ch.rings[service] = ring
	ch.ringToBackend[service] = nodeMap
}

func (ch *ConsistentHash) NextBackend(service string, r *http.Request) (string, error) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	ring := ch.rings[service]
	if len(ring) == 0 {
		return "", ErrNoBackends
	}

	key := ""
	if userID, _ := r.Context().Value(internalmiddleware.UserIDKey).(string); userID != "" {
		key = userID
	} else if hdr := r.Header.Get("X-User-ID"); hdr != "" {
		key = hdr
	} else {
		key = r.URL.Path
	}

	h := sha1.Sum([]byte(key))
	hash := binary.BigEndian.Uint32(h[:4])

	idx := sort.Search(len(ring), func(i int) bool { return ring[i] >= hash })
	if idx == len(ring) {
		idx = 0
	}

	backend := ch.ringToBackend[service][ring[idx]]
	if backend == "" {
		return "", ErrNoBackends
	}
	log.Infof("consistent hash key=%q service=%s backend=%s", key, service, backend)
	return backend, nil
}

func (ch *ConsistentHash) CurrentCount(service, backend string) int {
	return 0
}

func (rr *ConsistentHash) Done(service, backend string) {}