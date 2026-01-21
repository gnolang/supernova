# Supernova

A stress-testing tool for Gno TM2 networks. Simulates realistic transaction patterns and measures performance metrics like TPS and block utilization.

## Quick Start

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

For production-grade testing, increase `-sub-accounts` (50-100) and `-transactions` (5000+).

**Note**: This mnemonic derives the default development gnoland account, which is pre-funded in local environments (e.g., `gnodev`).

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

## Testing Modes

| Mode | What it Does | Best For |
|------|--------------|----------|
| REALM_DEPLOYMENT | Deploys a new realm per transaction | Testing heavy workloads (compilation, storage, state init) |
| PACKAGE_DEPLOYMENT | Deploys pure packages (stateless libraries) | Testing code storage overhead |
| REALM_CALL | Deploys one realm, then calls its methods | Simulating production workloads |

For most production scenarios, **REALM_CALL** provides the most relevant metrics since it simulates typical user interactions.

## Understanding Results

### TPS (Transactions Per Second)

Reflects real-world throughput, accounting for transaction propagation, block production intervals, and consensus overhead.

### Block Utilization

| Utilization | Meaning |
|-------------|---------|
| Low (<50%) | Network has spare capacity |
| High (>80%) | Near capacity, consider increasing gas limits |
| Variable | Inconsistent batching or congestion patterns |

## When to Use

- **Before deployment**: Validate network can handle expected load
- **After config changes**: Verify block gas limits, timing parameters
- **Capacity planning**: Determine hardware requirements for target TPS
- **Benchmarking**: Compare different node setups objectively

## Resources

- [Benchmark reports](https://github.com/gnolang/benchmarks/tree/main/reports/supernova)
