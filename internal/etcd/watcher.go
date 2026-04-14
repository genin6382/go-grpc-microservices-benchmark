package etcd

// import (
// 	"context"
// 	"strings"

// 	"github.com/genin6382/go-grpc-microservices-benchmark/gateway/loadbalancer"
// 	clientv3 "go.etcd.io/etcd/client/v3"
// 	log "github.com/sirupsen/logrus"
// )

// func WatchService(ctx context.Context, client *clientv3.Client, serviceName string, lb loadbalancer.LoadBalancer) error {
// 	prefix := "/services/" + serviceName + "/"

// 	resp, err := client.Get(ctx, prefix, clientv3.WithPrefix())
// 	if err != nil {
// 		return err
// 	}

// 	for _, kv := range resp.Kvs {
// 		lb.Add(string(kv.Value))
// 		log.Infof("Discovered existing %s endpoint: %s", serviceName, string(kv.Value))
// 	}

// 	wch := client.Watch(ctx, prefix, clientv3.WithPrefix())

// 	go func() {
// 		for {
// 			select {
// 			case wr, ok := <-wch:
// 				if !ok {
// 					log.Warnf("watch closed for %s", serviceName)
// 					return
// 				}

// 				for _, ev := range wr.Events {
// 					switch ev.Type {
// 					case clientv3.EventTypePut:
// 						addr := string(ev.Kv.Value)
// 						lb.Add(addr)
// 						log.Infof("Added %s endpoint: %s", serviceName, addr)

// 					case clientv3.EventTypeDelete:
// 						parts := strings.Split(string(ev.Kv.Key), "/")
// 						addr := parts[len(parts)-1]
// 						lb.Remove(addr)
// 						log.Infof("Removed %s endpoint: %s", serviceName, addr)
// 					}
// 				}

// 			case <-ctx.Done():
// 				log.Infof("stopping watcher for %s", serviceName)
// 				return
// 			}
// 		}
// 	}()

// 	return nil
// }