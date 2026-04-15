package gateway 

import (
	"net/http/httputil"
	"net/url"
	log "github.com/sirupsen/logrus"
	internalmiddleware "github.com/genin6382/go-grpc-microservices-benchmark/internal/middleware"
	"github.com/genin6382/go-grpc-microservices-benchmark/gateway/loadbalancer"
	"net/http"
)

func WithIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(internalmiddleware.UserIDKey).(string)
		if userID == "" {
			http.Error(w, "missing authenticated user", http.StatusUnauthorized)
			return
		}

		r.Header.Del("X-User-ID")
		r.Header.Set("X-User-ID", userID)

		next.ServeHTTP(w, r)
	})
}

func NewReverseProxy(target string) *httputil.ReverseProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("invalid target url %s: %v", target, err)
	}

	return &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(targetURL)
			pr.SetXForwarded()
			pr.Out.URL.Path = pr.In.URL.Path
			pr.Out.URL.RawPath = pr.In.URL.RawPath
			pr.Out.URL.RawQuery = pr.In.URL.RawQuery
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("proxy error for %s: %v", r.URL.Path, err)
			http.Error(w, "bad gateway", http.StatusBadGateway)
		},
		ModifyResponse: func(resp *http.Response) error {
			return nil
		},
	}
}

func ProxyHandler(lb loadbalancer.LoadBalancer, service string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backend, err := lb.NextBackend(service, r)
		if err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		if lc, ok := lb.(interface {
			Done(string, string)
			CurrentCount(string, string) int
		}); ok {
			log.Infof("Forwarding %s request to %s (active requests: %d)", service, backend, lc.CurrentCount(service, backend))
			defer lc.Done(service, backend)
		} else {
			log.Infof("Forwarding %s request to %s", service, backend)
		}

		proxy := NewReverseProxy(backend)
		proxy.ServeHTTP(w, r)
	})
}

