# Supernova

A stress-testing tool for Gno TM2 networks. Simulates realistic transaction
patterns and measures performance metrics like TPS and block utilization.
Use it to benchmark node performance by executing transactions under load
and analyzing response times.

## Table of Contents

- [Key Features](#key-features)
- [Why Stress Test Your Network?](#why-stress-test-your-network)
- [Quick Start](#quick-start)
- [CLI Flags](#cli-flags)
- [How Supernova Works](#how-supernova-works)
- [Testing Modes](#testing-modes)
- [Understanding Results](#understanding-results)
- [Integration with Monitoring](#integration-with-monitoring)
- [When to Use](#when-to-use)
- [Resources](#resources)

## Key Features

- Batch transactions for efficient stress testing
- Multiple modes: REALM_DEPLOYMENT, PACKAGE_DEPLOYMENT, REALM_CALL
- Distributed testing through subaccounts
- Automatic subaccount funding
- Detailed statistics and JSON output

To view the results of the stress tests, visit the [benchmarks reports for supernova](https://github.com/gnolang/benchmarks/tree/main/reports/supernova).

![Banner](.github/demo.gif)

## Why Stress Test Your Network?

Stress testing answers critical questions before production deployment:

| Question | Why It Matters |
|----------|----------------|
| What's my maximum TPS? | Know when performance degrades under load |
| How do gas limits affect throughput? | Optimize block parameters |
| Where are the bottlenecks? | Find issues in consensus, storage, or network |
| How do workloads differ? | Plan capacity for deployments vs calls |

## Quick Start

Requires Go 1.19 or higher.

```bash
# Build
make build

# Run a stress test
./build/supernova \
  -url http://localhost:26657 \
  -chain-id dev \
  -mnemonic "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast" \
  -sub-accounts 5 \
  -transactions 100 \
  -mode REALM_CALL \
  -output results.json
```

This runs a stress test against a Gno TM2 node at `http://localhost:26657`,
using `5` sub-accounts to send `100` transactions.
Results are saved to `results.json`.

For production-grade testing, increase `-sub-accounts` (50-100) and
`-transactions` (5000+).

**Note**: This mnemonic derives the default development gnoland account, which is pre-funded in local environments (e.g., `gnodev`). For other environments, ensure the first address (index 0) has sufficient funds for distribution to subaccounts.

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-url` | (required) | JSON-RPC URL of the Gno node |
| `-mnemonic` | (required) | Mnemonic for deriving accounts |
| `-sub-accounts` | 10 | Number of accounts sending transactions |
| `-transactions` | 100 | Total transactions to send |
| `-mode` | REALM_DEPLOYMENT | REALM_DEPLOYMENT, PACKAGE_DEPLOYMENT, or REALM_CALL |
| `-batch` | 100 | Batch size for JSON-RPC calls |
| `-chain-id` | dev | Chain ID of the network |
| `-output` | (none) | Path to save results JSON |

## How Supernova Works

Supernova operates in three phases:

### 1. Account Setup

Supernova derives multiple subaccounts from a single mnemonic. This
simulates realistic conditions where transactions come from many different
addresses, testing the network's ability to handle concurrent account
state updates.

The tool automatically distributes funds from the main account (index 0)
to all subaccounts before the test begins.

### 2. Transaction Generation

Based on the selected mode, supernova constructs and signs transactions.
See [Testing Modes](#testing-modes) for details on each mode.

### 3. Result Collection

After broadcasting transactions, supernova monitors the blockchain to
collect metrics like TPS, block utilization, and gas consumption.

## Testing Modes

| Mode | What it Does | Best For |
|------|--------------|----------|
| REALM_DEPLOYMENT | Deploys a new realm per transaction | Testing heavy workloads (compilation, storage, state init) |
| PACKAGE_DEPLOYMENT | Deploys pure packages (stateless libraries) | Testing code storage overhead |
| REALM_CALL | Deploys one realm, then calls its methods | Simulating production workloads |

For most production scenarios, **REALM_CALL** provides the most relevant
metrics since it simulates typical user interactions.

## Understanding Results

### TPS (Transactions Per Second)

Reflects real-world throughput, accounting for transaction propagation, block
production intervals, and consensus overhead.

### Block Utilization

| Utilization | Meaning |
|-------------|---------|
| Low (<50%) | Network has spare capacity |
| High (>80%) | Near capacity, consider increasing gas limits |
| Variable | Inconsistent batching or congestion patterns |

## Integration with Monitoring

For deeper insights, run supernova against a node with
[OpenTelemetry enabled](https://github.com/gnolang/gno/tree/master/misc/telemetry).
This allows you to correlate supernova's transaction metrics with internal
node metrics like:

- Memory and CPU usage during load
- Consensus round timing
- Storage I/O patterns
- Network message latency

## When to Use

| Scenario | Purpose |
|----------|---------|
| Pre-production | Validate network handles expected load |
| Post-config changes | Verify gas limits and timing parameters |
| Capacity planning | Determine hardware for target TPS |
| Benchmarking | Compare node configurations |

## Resources

- [Benchmark reports](https://github.com/gnolang/benchmarks/tree/main/reports/supernova)
- [Gno.land documentation](https://docs.gno.land)
