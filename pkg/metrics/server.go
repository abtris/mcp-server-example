package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server provides an HTTP server for Prometheus metrics
type Server struct {
	addr   string
	server *http.Server
}

// NewServer creates a new metrics HTTP server
func NewServer(port int) *Server {
	addr := fmt.Sprintf(":%d", port)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &Server{
		addr: addr,
		server: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

// Start starts the metrics HTTP server
func (s *Server) Start() error {
	slog.Info("Starting metrics server", "addr", s.addr, "endpoint", "/metrics")
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("metrics server error: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the metrics server
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down metrics server")
	return s.server.Shutdown(ctx)
}

