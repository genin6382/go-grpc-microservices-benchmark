package etcd

import (
	"context"
	"fmt"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
	log "github.com/sirupsen/logrus"

	"github.com/genin6382/go-grpc-microservices-benchmark/gateway/loadbalancer"
)

func normalizeAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	return "http://" + addr
}

func WatchService(ctx context.Context, client *clientv3.Client, serviceName string, lb loadbalancer.LoadBalancer) error {
	prefix := fmt.Sprintf("/services/%s/", serviceName)

	resp, err := client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	backends := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		addr := normalizeAddr(string(kv.Value))
		if addr == "" {
			continue
		}
		backends = append(backends, addr)
		log.Infof("Discovered existing %s endpoint: %s", serviceName, addr)
	}
	lb.UpdateBackends(serviceName, backends)

	wch := client.Watch(ctx, prefix, clientv3.WithPrefix())

	go func() {
		for {
			select {
			case wr, ok := <-wch:
				if !ok {
					log.Warnf("watch closed for %s", serviceName)
					return
				}

				current, err := client.Get(ctx, prefix, clientv3.WithPrefix())
				if err != nil {
					log.Errorf("failed to refresh backends for %s: %v", serviceName, err)
					continue
				}

				updated := make([]string, 0, len(current.Kvs))
				seen := make(map[string]struct{})

				for _, kv := range current.Kvs {
					addr := normalizeAddr(string(kv.Value))
					if addr == "" {
						continue
					}
					if _, ok := seen[addr]; ok {
						continue
					}
					seen[addr] = struct{}{}
					updated = append(updated, addr)
				}

				lb.UpdateBackends(serviceName, updated)
				log.Infof("Updated %s backends: %v", serviceName, updated)

				_ = wr
			case <-ctx.Done():
				log.Infof("stopping watcher for %s", serviceName)
				return
			}
		}
	}()

	return nil
}