package internal

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

var ErrNoAvailableServer = errors.New("No available server, try again later")

type LoadBalancer struct {
	counter           *atomic.Uint64
	servers           []*DowmstreamServer
	downstreamClient  *http.Client
	healthCheckClient *http.Client
	healthCheckTime   *time.Ticker
	healthCheckDone   chan bool
}

func NewLoadBalancer(healthCheckTimeout, requestTimeout, healthCheckInterval int, servers []*DowmstreamServer) *LoadBalancer {
	ticker := time.NewTicker(time.Duration(healthCheckInterval) * time.Millisecond)

	dc := &http.Client{
		Timeout: time.Duration(requestTimeout) * time.Millisecond,
	}

	hc := &http.Client{
		Timeout: time.Duration(healthCheckTimeout) * time.Millisecond,
	}

	return &LoadBalancer{
		counter:           new(atomic.Uint64),
		servers:           servers,
		downstreamClient:  dc,
		healthCheckClient: hc,
		healthCheckTime:   ticker,
		healthCheckDone:   make(chan bool),
	}
}

func (l *LoadBalancer) Shutdown() {
	l.healthCheckDone <- true
}

func (l *LoadBalancer) HealthCheck() {
	for {
		select {
		case <-l.healthCheckDone:
			return
		case <-l.healthCheckTime.C:
			l.healthCheck()
		}
	}
}

func (l *LoadBalancer) healthCheck() {
	for _, server := range l.servers {
		go func(ds *DowmstreamServer) {
			_, err := l.healthCheckClient.Get(ds.healthCheckAddress)
			if err != nil {
				ds.Unhealthy()
			} else {
				ds.Healthy()
			}
			slog.Info("HealthCheck", "server", ds.healthCheckAddress, "healthy", ds.healthy)
		}(server)
	}
}

func (l *LoadBalancer) RoundRobin() *DowmstreamServer {
	server := l.servers[int(l.counter.Load())%len(l.servers)]
	l.counter.Add(1)
	return server
}

func (l *LoadBalancer) GetDownstreamServer() (*DowmstreamServer, error) {
	retry := 0
	server := l.RoundRobin()
	for !server.healthy {
		if retry == len(l.servers)-1 {
			return nil, ErrNoAvailableServer
		}
		server = l.RoundRobin()
		retry++
	}
	return server, nil
}

func (l *LoadBalancer) ErrorResponse(message string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(message))
}

func (l *LoadBalancer) NewRequest(server *DowmstreamServer, r *http.Request) (*http.Request, error) {
	req, err := http.NewRequest(r.Method, server.address, nil)
	if err != nil {
		return nil, err
	}

	req.Header = r.Header
	req.Host = r.Host
	req.RemoteAddr = r.RemoteAddr
	return req, nil
}

func logRequest(server *DowmstreamServer, r *http.Request) {
	slog.Info("Request",
		"origin", r.RemoteAddr,
		"method", r.Method,
		"url", r.URL,
		"host", r.Host,
		"user-agent", r.UserAgent(),
		"destination", server.address)
}

func (l *LoadBalancer) Handle(w http.ResponseWriter, r *http.Request) {
	for {
		server, err := l.GetDownstreamServer()
		if err != nil {
			l.ErrorResponse(err.Error(), w, r)
			return
		}

		req, err := l.NewRequest(server, r)
		if err != nil {
			l.ErrorResponse(err.Error(), w, r)
			return
		}

		resp, err := l.downstreamClient.Do(req)
		if err != nil {
			server.Unhealthy()
			continue
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			l.ErrorResponse(err.Error(), w, r)
			return
		}

		logRequest(server, r)
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		break
	}
}
