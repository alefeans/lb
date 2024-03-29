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

func TestNewDownstreamServers(t *testing.T) {
	tests := []struct {
		input    string
		endpoint string
		want     []string
	}{
		{input: "http://localhost", endpoint: "/", want: []string{"http://localhost/"}},
		{input: "http://localhost:80,http://localhost:81", endpoint: "/", want: []string{"http://localhost:80/", "http://localhost:81/"}},
		{input: "http://localhost:80,http://localhost:81,http://localhost:82", endpoint: "/", want: []string{"http://localhost:80/", "http://localhost:81/", "http://localhost:82/"}},
	}

	for _, test := range tests {
		got := NewDownstreamServers(test.input, test.endpoint)

		if len(got) != len(test.want) {
			t.Errorf("got %d, want %d", len(got), len(test.want))
		}

		for i := 0; i < len(got); i++ {
			if got[i].healthCheckAddress != test.want[i] {
				t.Errorf("got %s, want %s", got[i].healthCheckAddress, test.want[i])
			}
		}
	}
}
