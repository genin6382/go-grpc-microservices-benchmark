package etcd

import (
	"context"
	"fmt"
	
	log "github.com/sirupsen/logrus"
)

func (s *ServiceRegistry) Deregister(ctx context.Context, serviceName, serviceAddr string) error {
	key := fmt.Sprintf("/services/%s/%s", serviceName, serviceAddr)
	_, err := s.client.Delete(ctx, key)
	if err != nil {
		return err
	}
	log.Infof("Deregistered service %s at %s", serviceName, serviceAddr)
	return nil
}
