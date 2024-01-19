package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
)

type StubDownstreamServer struct {
	h *http.Server
}

func NewStubDownstreamServer(addr string) *StubDownstreamServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("Hello from %s", addr)
		w.Write([]byte(msg))
	})
	return &StubDownstreamServer{
		h: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

func (s *StubDownstreamServer) Start() {
	s.h.ListenAndServe()
}

func (s *StubDownstreamServer) Shutdown() {
	if err := s.h.Shutdown(context.Background()); err != nil {
		return
	}
}

func TestLoadBalancing(t *testing.T) {
	stub1 := NewStubDownstreamServer(":8081")
	go stub1.Start()
	defer stub1.Shutdown()

	stub2 := NewStubDownstreamServer(":8082")
	go stub2.Start()
	defer stub2.Shutdown()

	stub3 := NewStubDownstreamServer(":8083")
	go stub3.Start()
	defer stub3.Shutdown()

	downstreamServers := NewDownstreamServers("http://localhost:8081,http://localhost:8082,http://localhost:8083", "/")
	lb := NewLoadBalancer(100, 100, 100, downstreamServers)
	server := NewServer(":80", lb)
	go server.Start()
	defer server.Shutdown(context.Background())


	client := &http.Client{}
	wants := []string{"Hello from :8081", "Hello from :8082", "Hello from :8083", "Hello from :8081", "Hello from :8082", "Hello from :8083"}

	for _, want := range wants {
		resp, err := client.Get("http://localhost:80/")
		if err != nil {
			t.Errorf("got unexpected error: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("got %d, want 200", resp.StatusCode)
		}

		defer resp.Body.Close()
		got, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("got unexpected error: %v", err)
		}

		if string(got) != want {
			t.Errorf("got %s, want %s", got, want)
		}
	}
}
