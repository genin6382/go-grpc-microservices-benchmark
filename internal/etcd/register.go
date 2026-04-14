package etcd

import (
	"fmt"
	"time"

	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
	log "github.com/sirupsen/logrus"
)

type ServiceRegistry struct {
	client *clientv3.Client
	leaseID clientv3.LeaseID
}

func NewServiceRegistry(endpoints []string)(*ServiceRegistry, error) {
	cli , err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
		DialTimeout: 5*time.Second,
	})

	if err != nil {
		return nil , err
	}
	return &ServiceRegistry{client: cli}, nil
}

func (s *ServiceRegistry) Register(ctx context.Context, serviceName, serviceAddr string, ttl int64) error {
	leaseResp, err := s.client.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("/services/%s/%s", serviceName, serviceAddr)
	_, err = s.client.Put(ctx, key, serviceAddr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	keepAliveChan, err := s.client.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case ka, ok := <-keepAliveChan:
				if !ok || ka == nil {
					log.Warnf("Lease expired for %s", serviceName)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	log.Infof("Registered service %s at %s with TTL %d", serviceName, serviceAddr, ttl)

	return nil
}