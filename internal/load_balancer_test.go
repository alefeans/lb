package internal

import (
	"testing"
	"time"
)

func LoadBalancerWithValidServer() *LoadBalancer {
	server := NewDownstreamServer("http://localhost:8081", "/")
	return NewLoadBalancer(100, 100, 100, []*DowmstreamServer{server})
}

func LoadBalancerWithInvalidServer() *LoadBalancer {
	server := NewDownstreamServer("invalid", "/")
	return NewLoadBalancer(1, 1, 1, []*DowmstreamServer{server})
}

func TestHealthCheck(t *testing.T) {
	t.Run("Invalid Server", func(t *testing.T) {
		lb := LoadBalancerWithInvalidServer()
		go lb.HealthCheck()
		<-time.After(time.Millisecond * 105)

		got := lb.servers[0].healthy
		if got {
			t.Errorf("got %t, want false", got)
		}
	})

	t.Run("Valid Server", func(t *testing.T) {
		stub := NewStubDownstreamServer(":8081")
		go stub.Start()
		defer stub.Shutdown()

		lb := LoadBalancerWithValidServer()
		go lb.HealthCheck()
		<-time.After(time.Millisecond * 105)

		got := lb.servers[0].healthy
		if !got {
			t.Errorf("got %t, want true", got)
		}
	})

	t.Run("Healthy Server Becoming Unhealthy)", func(t *testing.T) {
		lb := LoadBalancerWithValidServer()
		go lb.HealthCheck()
		<-time.After(time.Millisecond * 105)

		got := lb.servers[0].healthy
		if got {
			t.Errorf("got %t, want false", got)
		}
	})

	t.Run("Unhealthy Server Becoming Healthy", func(t *testing.T) {
		lb := LoadBalancerWithValidServer()
		go lb.HealthCheck()
		<-time.After(time.Millisecond * 105)

		got := lb.servers[0].healthy
		if got {
			t.Errorf("got %t, want false", got)
		}

		stub := NewStubDownstreamServer(":8081")
		go stub.Start()
		defer stub.Shutdown()
		<-time.After(time.Millisecond * 105)

		got = lb.servers[0].healthy
		if !got {
			t.Errorf("got %t, want true", got)
		}
	})

	t.Run("Shutdown", func(t *testing.T) {
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
