package loadbalancer

import (
	"errors"
	"net/http"
)

var ErrNoBackends = errors.New("no backends available")

type LoadBalancer interface {
	UpdateBackends(service string, backends []string)
	NextBackend(service string, r *http.Request) (string, error)
	Done(service string, backend string)
	CurrentCount(service string, backend string) int
}