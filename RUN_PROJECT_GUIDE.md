# Running the ERST Project with Prometheus Metrics

This guide explains how to build and run the ERST project with the new Prometheus metrics functionality.

## Prerequisites

### Required
- **Go 1.21+** (1.23 recommended)
- **Rust 1.85.0+** (for simulator)
- **Git**

### Optional
- **Docker** (for Prometheus/Grafana)
- **Make** (for using Makefile commands)

## Installation

### 1. Install Go

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install golang-go
```

**macOS:**
```bash
brew install go
```

**Or download from:** https://golang.org/dl/

Verify installation:
```bash
go version
# Should show: go version go1.23.x ...
```

### 2. Install Rust (if not installed)

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
```

Verify:
```bash
rustc --version
cargo --version
```

## Building the Project

### Step 1: Fix Dependencies

First, update `go.sum` to fix the CI issue:

```bash
# Run the fix script
./fix_ci.sh

# Or manually
go mod tidy
go mod verify
```

This will download the Prometheus dependency and update `go.sum`.

### Step 2: Build the Rust Simulator

```bash
cd simulator
cargo build --release
cd ..
```

This creates `simulator/target/release/erst-sim`.

### Step 3: Build the Go CLI

```bash
# Using Make (recommended)
make build

# Or directly with go
go build -o bin/erst ./cmd/erst
```

This creates `bin/erst`.

### Step 4: Verify Build

```bash
./bin/erst --version
```

Expected output:
```
erst version dev (commit: <sha>, built: <date>)
```

## Running the Daemon with Metrics

### Start the Daemon

```bash
./bin/erst daemon --port 8080 --network testnet
```

Expected output:
```
INFO Starting JSON-RPC server port=8080
```

### Verify Metrics Endpoint

In another terminal:

```bash
# Check health
curl http://localhost:8080/health
# Expected: {"status":"ok"}

# Check metrics
curl http://localhost:8080/metrics
```

Expected metrics output:
```
# HELP remote_node_last_response_timestamp_seconds Unix timestamp of the last successful simulation response from a remote node
# TYPE remote_node_last_response_timestamp_seconds gauge
# HELP remote_node_response_duration_seconds Duration of simulation requests to remote nodes in seconds
# TYPE remote_node_response_duration_seconds histogram
# HELP remote_node_response_total Total number of simulation responses from remote nodes by status
# TYPE remote_node_response_total counter
# HELP simulation_execution_total Total number of simulation executions by status
# TYPE simulation_execution_total counter
```

### Trigger a Simulation

To generate metrics data, run a simulation:

```bash
# Debug a transaction (replace with real hash)
./bin/erst debug <transaction_hash> --network testnet

# Or use the JSON-RPC API
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "DebugTransaction",
    "params": {"hash": "<transaction_hash>"},
    "id": 1
  }'
```

### View Updated Metrics

```bash
curl http://localhost:8080/metrics | grep remote_node
```

You should now see actual metric values:
```
remote_node_last_response_timestamp_seconds{network="testnet",node_address="https://horizon-testnet.stellar.org/"} 1.709123456e+09
remote_node_response_total{network="testnet",node_address="https://horizon-testnet.stellar.org/",status="success"} 1
remote_node_response_duration_seconds_count{network="testnet",node_address="https://horizon-testnet.stellar.org/"} 1
```

## Running Tests

### Unit Tests

```bash
# Test metrics package
go test ./internal/metrics -v

# Test all packages
go test ./... -v

# With race detection
go test -race ./...
```

### Integration Tests

```bash
# Run integration tests (requires build tag)
go test -tags=integration ./internal/metrics -v
```

### Benchmarks

```bash
# Run all benchmarks
make bench

# Run specific benchmarks
go test -bench=. -benchmem ./internal/metrics
```

## Setting Up Prometheus

### 1. Create Prometheus Config

Create `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'erst-daemon'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### 2. Run Prometheus with Docker

```bash
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus
```

### 3. Access Prometheus UI

Open http://localhost:9090

Try these queries:
```promql
# Check staleness
time() - remote_node_last_response_timestamp_seconds

# Error rate
rate(remote_node_response_total{status="error"}[5m]) / rate(remote_node_response_total[5m])

# p95 latency
histogram_quantile(0.95, rate(remote_node_response_duration_seconds_bucket[5m]))
```

## Setting Up Grafana

### 1. Run Grafana with Docker

```bash
docker run -d \
  --name grafana \
  -p 3000:3000 \
  grafana/grafana
```

### 2. Access Grafana

1. Open http://localhost:3000
2. Login: admin/admin
3. Add Prometheus data source:
   - URL: http://host.docker.internal:9090 (Mac/Windows)
   - URL: http://172.17.0.1:9090 (Linux)

### 3. Create Dashboard

Import or create panels with queries from `docs/PROMETHEUS_METRICS.md`.

## Development Workflow

### 1. Make Code Changes

Edit files in `internal/metrics/` or other packages.

### 2. Format Code

```bash
make fmt
```

### 3. Run Linters

```bash
make lint-strict
```

### 4. Run Tests

```bash
go test ./...
```

### 5. Build

```bash
make build
```

### 6. Test Locally

```bash
./bin/erst daemon --port 8080 --network testnet
```

## Troubleshooting

### "Go not found"

Install Go from https://golang.org/dl/

### "erst-sim not found"

Build the Rust simulator:
```bash
cd simulator && cargo build --release
```

### "go.sum mismatch"

Run:
```bash
go mod tidy
go mod verify
```

### "Port 8080 already in use"

Use a different port:
```bash
./bin/erst daemon --port 8081 --network testnet
```

### "No metrics data"

Trigger a simulation to generate metrics:
```bash
./bin/erst debug <transaction_hash> --network testnet
```

### "Prometheus can't scrape metrics"

Check:
1. Daemon is running: `curl http://localhost:8080/health`
2. Metrics endpoint works: `curl http://localhost:8080/metrics`
3. Prometheus config has correct target
4. No firewall blocking port 8080

## Quick Start Commands

```bash
# Complete setup from scratch
go mod tidy                                    # Fix dependencies
make build                                     # Build CLI
./bin/erst daemon --port 8080 --network testnet  # Start daemon
curl http://localhost:8080/metrics            # Check metrics

# In another terminal
./bin/erst debug <tx_hash> --network testnet  # Generate metrics
curl http://localhost:8080/metrics | grep remote_node  # View metrics
```

## Environment Variables

```bash
# Simulator path
export ERST_SIM_PATH=/path/to/erst-sim

# Telemetry
export ERST_TELEMETRY_ENABLED=true
export ERST_OTLP_ENDPOINT=http://localhost:4318

# Logging
export ERST_LOG_LEVEL=debug
```

## Docker Compose (All-in-One)

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

Run:
```bash
docker-compose up -d
```

## Performance Tips

### 1. Use Release Build

```bash
make build-release
```

### 2. Enable Caching

```bash
export GOCACHE=$(go env GOCACHE)
```

### 3. Parallel Tests

```bash
go test -parallel 4 ./...
```

## Next Steps

1. **Read Documentation**
   - `docs/PROMETHEUS_METRICS.md` - Full metrics guide
   - `docs/METRICS_VERIFICATION.md` - Verification steps
   - `docs/METRICS_QUICK_REFERENCE.md` - Quick reference

2. **Set Up Monitoring**
   - Configure Prometheus scraping
   - Create Grafana dashboards
   - Set up alerting rules

3. **Integrate with CI/CD**
   - Add metrics tests to pipeline
   - Monitor deployment health
   - Track performance over time

## Support

For issues or questions:
- Check `docs/METRICS_TESTING.md` for testing guide
- Check `CI_FAILURE_ANALYSIS.md` for CI issues
- Review `IMPLEMENTATION_STATUS.md` for status

## Summary

```bash
# Quick start (requires Go and Rust)
go mod tidy                    # Fix dependencies
make build                     # Build everything
./bin/erst daemon --port 8080  # Start with metrics
curl http://localhost:8080/metrics  # View metrics
```

The metrics are now live at `/metrics` endpoint! 🎉
