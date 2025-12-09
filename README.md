# URL Status Checker

[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/url-status-checker)](https://goreportcard.com/report/github.com/yourusername/url-status-checker)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance, concurrent URL status checker built with idiomatic Go. Check thousands of URLs in parallel with goroutines, get detailed response metrics, and monitor via REST API, CLI, or web dashboard.

## Features

🚀 **High Concurrency**: Leverages goroutines, channels, and WaitGroups for efficient parallel checking
🌐 **Multiple Interfaces**: REST API, CLI, and web dashboard
📊 **Prometheus Metrics**: Export metrics for monitoring and alerting
🐳 **Docker Ready**: Single binary or containerized deployment
⚡ **Fast & Lightweight**: Sub-second responses, minimal memory footprint
🧪 **Production Ready**: Comprehensive tests, linting, error handling

## Quick Start
### Clone repo
```bash
git clone https://github.com/tluolamo/url-status-checker.git
cd url-status-checker
```

### Running Locally (live reload)

```bash
task dev
```

### Using Docker (Recommended)

```bash
task docker-run
```

### Using Kubernetes (local demo)

```bash
task kind-demo
```

## Usage

### REST API

Check URLs via API:
```bash
curl -X POST http://localhost:8080/api/v1/check \
  -H "Content-Type: application/json" \
  -d '{
    "urls": ["https://google.com", "https://github.com"],
    "timeout": "5s",
    "max_workers": 10
  }'
```

Response:
```json
{
  "results": [
    {
      "url": "https://google.com",
      "status_code": 200,
      "response_time_ms": 145,
      "available": true,
      "error": null
    },
    {
      "url": "https://github.com",
      "status_code": 200,
      "response_time_ms": 234,
      "available": true,
      "error": null
    }
  ],
  "total_checked": 2,
  "total_available": 2,
  "total_time_ms": 250
}
```

### Web Dashboard

Open your browser to `http://localhost:8080` to access the interactive dashboard.

### Prometheus Metrics

Metrics are exposed at `http://localhost:8080/metrics`:

```
# HELP url_checks_total Total number of URL checks performed
# TYPE url_checks_total counter
url_checks_total{status="success"} 1523
url_checks_total{status="failure"} 47

# HELP url_check_duration_seconds Time taken to check URLs
# TYPE url_check_duration_seconds histogram
url_check_duration_seconds_bucket{le="0.1"} 890
url_check_duration_seconds_bucket{le="0.5"} 1450
```

## Configuration

Configuration via environment variables or CLI flags:

| Environment Variable | CLI Flag | Default | Description |
|---------------------|----------|---------|-------------|
| `PORT` | `--port` | `8080` | HTTP server port |
| `MAX_WORKERS` | `--workers` | `100` | Max concurrent workers |
| `DEFAULT_TIMEOUT` | `--timeout` | `10s` | Default request timeout |
| `LOG_LEVEL` | `--log-level` | `info` | Logging level (debug, info, warn, error) |

## Development

### Prerequisites

- Go 1.22 or later
- Task
- Docker (optional)

### Setup

```bash
# Install dependencies
task deps

# Run tests
task test

# Run linter
task lint

# Run with hot reload
task dev
```

### Project Structure

```
url-status-checker/
├── cmd/
│   └── urlchecker/          # Application entry point
├── internal/
│   ├── api/                 # HTTP handlers
│   ├── checker/             # Core URL checking logic
│   ├── config/              # Configuration management
│   ├── metrics/             # Prometheus metrics
│   └── models/              # Data models
├── deployments/             # Docker and deployment configs
├── bin/                     # Compiled binaries
└── tmp/                     # Temporary files (e.g., for hot reload)
```

### Running Tests

```bash
# Unit tests
task test

# Coverage report
task coverage

# Benchmarks
task bench

# Race detector
task test-race
```

## Docker Deployment

### Docker Compose with Monitoring Stack

```bash
task docker-compose-up
```

This starts:
- URL Checker (port 8080)
- Prometheus (port 9090)
- Grafana (port 3000)

### Kubernetes (Local Demo)

Run the complete demo setup with one command:

```bash
task kind-demo
```

This will:
1. Create a kind (Kubernetes in Docker) cluster
2. Build and load the Docker image
3. Deploy the application and Prometheus
4. Access at http://localhost:8080

For detailed Kubernetes deployment instructions, see [deployments/kubernetes/README.md](deployments/kubernetes/README.md).

**Prerequisites**: Install [kind](https://kind.sigs.k8s.io/) and [kubectl](https://kubernetes.io/docs/tasks/tools/)

**Management commands**:
```bash
task k8s-status      # Check deployment status
task k8s-logs        # View application logs
task k8s-delete      # Remove deployment
task kind-delete     # Delete cluster
```

## Performance

- **Throughput**: 1000+ URLs per second
- **Memory**: < 50MB baseline
- **Binary Size**: < 10MB
- **Docker Image**: < 20MB (Alpine-based)

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with idiomatic Go to demonstrate:
- Goroutines and channels for concurrency
- Standard library HTTP server
- Clean architecture and separation of concerns
- Modern DevOps practices
