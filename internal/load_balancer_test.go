package internal

import (
	"context"
	"net/http"
	"testing"
	"time"
)

var ValidAddress = "http://localhost:8081"

type StubServer struct {
	h *http.Server
}

func NewStubServer() *StubServer {
	return &StubServer{
		h: &http.Server{
			Addr: ValidAddress,
		},
	}
}

func (s *StubServer) Start() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	s.h.ListenAndServe()
}

func (s *StubServer) Shutdown() {
	if err := s.h.Shutdown(context.Background()); err != nil {
		return
	}
}

func LoadBalancerWithValidServer() *LoadBalancer {
	server := NewDownstreamServer(ValidAddress, "/")
	return NewLoadBalancer(1, 1, 1, []*DowmstreamServer{server})
}

func LoadBalancerWithInvalidServer() *LoadBalancer {
	server := NewDownstreamServer("invalid", "/")
	return NewLoadBalancer(1, 1, 1, []*DowmstreamServer{server})
}

func TestHealthCheck(t *testing.T) {

	t.Run("HealthCheck Invalid Server", func(t *testing.T) {
		lb := LoadBalancerWithInvalidServer()
		go lb.HealthCheck()
		<-time.After(time.Millisecond * 1003)

		got := lb.servers[0].healthy
		if got {
			t.Errorf("got %t, want false", got)
		}
	})

	t.Run("HealthCheck Valid Server", func(t *testing.T) {
		stub := NewStubServer()
		go stub.Start()
		
		lb := LoadBalancerWithValidServer()
		go lb.HealthCheck()
		<-time.After(time.Millisecond * 1003)

		got := lb.servers[0].healthy
		if !got {
			t.Errorf("got %t, want true", got)
		}

		// Becoming Unavailable
		stub.Shutdown()
		<-time.After(time.Millisecond * 1003)

		got = lb.servers[0].healthy
		if got {
			t.Errorf("got %t, want false", got)
		}
	})

	t.Run("HealthCheck Shutdown", func(t *testing.T) {
		lb := LoadBalancerWithValidServer()
		go lb.HealthCheck()
		lb.Shutdown()

		got := lb.servers[0].healthy
		if !got {
			t.Errorf("got %t, want true", got)
		}
	})
}

func TestRoundRobin(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{input: "server1", want: []string{"server1", "server1"}},
		{input: "server1,server2", want: []string{"server1", "server2", "server1", "server2"}},
		{input: "server1,server2,server3", want: []string{"server1", "server2", "server3", "server1", "server2", "server3"}},
	}

	for _, test := range tests {
		servers := NewDownstreamServers(test.input, "/")
		lb := NewLoadBalancer(1, 1, 1, servers)
		for _, want := range test.want {
			if got := lb.RoundRobin().address; got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		}
	}
}

func TestGetDownstreamServer(t *testing.T) {
	t.Run("No Available Server", func(t *testing.T) {
		lb := LoadBalancerWithValidServer()
		lb.servers[0].Unhealthy()
		if got, err := lb.GetDownstreamServer(); got != nil && err != ErrNoAvailableServer {
			t.Errorf("got %v and err %v, want nil and ErrNoAvailableServer", got, err)
		}
	})

	t.Run("One Available Server", func(t *testing.T) {
		lb := LoadBalancerWithValidServer()
		want := lb.servers[0].address
		if got, err := lb.GetDownstreamServer(); got.address != want && err == ErrNoAvailableServer {
			t.Errorf("got %v and err %v, want %s and nil", got, err, want)
		}
	})
	
	t.Run("One Unavailable Server and One Available Server", func(t *testing.T) {
		lb := LoadBalancerWithValidServer()
		lb.servers[0].Unhealthy()
		lb.servers = append(lb.servers, NewDownstreamServer("available", "/"))
		want := lb.servers[1].address

		if got, err := lb.GetDownstreamServer(); got.address != want && err == ErrNoAvailableServer {
			t.Errorf("got %v and err %v, want %s and nil", got, err, want)
		}
	})
}