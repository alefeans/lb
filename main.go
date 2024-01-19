package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/alefeans/lb/internal"
)

type Args struct {
	address             string
	servers             string
	requestTimeout      int
	healthCheckEndpoint string
	healthCheckTimeout  int
	healthCheckInterval int
}

func parseCliArgs() *Args {
	var args Args
	flag.StringVar(&args.address, "a", ":80", "Load balancer server address")
	flag.IntVar(&args.requestTimeout, "r", 3000, "Request timeout for downstream servers in milliseconds")
	flag.IntVar(&args.healthCheckTimeout, "t", 3000, "Health check timeout in milliseconds")
	flag.IntVar(&args.healthCheckInterval, "i", 3000, "Downstream servers health check interval in milliseconds")
	flag.StringVar(&args.healthCheckEndpoint, "u", "/", `Health check endpoint (e.g "/health-check")`)
	flag.StringVar(&args.servers, "s", "", "Comma-separated list of downstream servers (e.g. http://0.0.0.0:8080,http://localhost:8081)")
	flag.Parse()
	return &args
}

func main() {
	args := parseCliArgs()
	if args.servers == "" {
		slog.Error("No downstream servers were set")
		flag.Usage()
		os.Exit(1)
	}

	servers := internal.NewDownstreamServers(args.servers, args.healthCheckEndpoint)
	lb := internal.NewLoadBalancer(args.healthCheckTimeout, args.requestTimeout, args.healthCheckInterval, servers)
	server := internal.NewServer(args.address, lb)
	go server.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.GracefulShutdown(ctx)
}
