package internal

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	lb         *LoadBalancer
	address    string
	httpServer *http.Server
}

func NewServer(address string, lb *LoadBalancer) *Server {
	return &Server{
		lb:      lb,
		address: address,
		httpServer: &http.Server{
			Addr: address,
		},
	}
}

func (s *Server) Start() {
	slog.Info("Starting Load Balancer server", "address", s.address)
	go s.lb.HealthCheck()
	http.HandleFunc("/", s.lb.Handle)
	s.httpServer.ListenAndServe()
}

func (s *Server) GracefulShutdown(ctx context.Context) {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	s.lb.Shutdown()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("Graceful shutdown failed", "error", err.Error())
		return
	}

	slog.Info("Graceful shutdown completed")
}
