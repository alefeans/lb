package internal

import (
	"fmt"
	"strings"
)

type DowmstreamServer struct {
	healthy            bool
	address            string
	healthCheckAddress string
}

func NewDownstreamServer(address, endpoint string) *DowmstreamServer {
	return &DowmstreamServer{
		healthy:            true,
		address:            address,
		healthCheckAddress: fmt.Sprintf("%s%s", address, endpoint),
	}
}

func (d *DowmstreamServer) Healthy() {
	d.healthy = true
}

func (d *DowmstreamServer) Unhealthy() {
	d.healthy = false
}

func NewDownstreamServers(argServers, healthCheckEndpoint string) []*DowmstreamServer {
	var servers []*DowmstreamServer
	for _, server := range strings.Split(argServers, ",") {
		servers = append(servers, NewDownstreamServer(server, healthCheckEndpoint))
	}
	return servers
}
