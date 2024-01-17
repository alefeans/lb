# lb

`lb` is a Go implementation of an HTTP Load Balancer that uses the [Round Robin](https://www.nginx.com/resources/glossary/round-robin-load-balancing/#:~:text=What%20Is%20Round%2DRobin%20Load,to%20each%20server%20in%20turn.) algorithm and the solution for the challenge [Write You Own Load Balancer
](https://codingchallenges.fyi/challenges/challenge-load-balancer/).

### Usage

```sh
go build

./lb -h
Usage of ./lb:
  -a string
        Load balancer server address (default ":80")
  -i int
        Downstream servers health check interval in seconds (default 10)
  -r int
        Request timeout for downstream servers in seconds (default 10)
  -s string
        Comma-separated list of downstream servers (e.g. http://0.0.0.0:8080,http://localhost:8081)
  -t int
        Health check timeout in seconds (default 10)
  -u string
        Health check endpoint (e.g "/health-check") (default "/")
```

The [be](/be/) directory contains some simple HTML files that can be used by web servers for testing purposes. See an example using Python's built-in web servers:

![](/docs/lb.gif)

### Tests

```sh
go test -v ./...
```

### Benchmarks

```sh
go test ./... -bench=. -benchmem

```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
