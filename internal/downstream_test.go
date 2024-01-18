package internal

import "testing"

func TestNewDownstreamServer(t *testing.T) {
	tests := []struct {
		address  string
		endpoint string
		want     string
	}{
		{address: "http://localhost", endpoint: "/", want: "http://localhost/"},
		{address: "http://localhost:80", endpoint: "/", want: "http://localhost:80/"},
		{address: "http://localhost:80", endpoint: "/test", want: "http://localhost:80/test"},
		{address: "http://localhost:80", endpoint: "/health/check", want: "http://localhost:80/health/check"},
	}

	for _, test := range tests {
		got := NewDownstreamServer(test.address, test.endpoint)
		if got.healthCheckAddress != test.want {
			t.Errorf("got %s, want %s", got.healthCheckAddress, test.want)
		}
		if !got.healthy {
			t.Errorf("got %t, want true", got.healthy)
		}
	}
}
