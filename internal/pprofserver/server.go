package pprofserver

import (
	"context"
	"crypto/subtle"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/cdxy1/go-courier-service/pkg/config"
)

type Server struct {
	httpServer *http.Server
	cfg        *config.PprofConfig
}

func New(cfg *config.PprofConfig) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           accessGuard(cfg, mux),
			ReadHeaderTimeout: 5 * time.Second,
		},
		cfg: cfg,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func accessGuard(cfg *config.PprofConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isLoopback(r.RemoteAddr) {
			next.ServeHTTP(w, r)
			return
		}
		if hasBasicAuth(cfg) {
			if validateBasicAuth(r, cfg) {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Add("WWW-Authenticate", `Basic realm="pprof"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusForbidden)
	})
}

func hasBasicAuth(cfg *config.PprofConfig) bool {
	return cfg != nil && cfg.BasicUser != "" && cfg.BasicPassword != ""
}

func validateBasicAuth(r *http.Request, cfg *config.PprofConfig) bool {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(user), []byte(cfg.BasicUser)) != 1 {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(pass), []byte(cfg.BasicPassword)) != 1 {
		return false
	}
	return true
}

func isLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback()
}
