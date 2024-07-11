<div align="center">
  <h1> Tokamak Trunks</h1>
</div>

## TL;DR

This is load test tool for tokamak-network L2 chain

Implemented using [vegeta](https://github.com/tsenart/vegeta)

## Introduction

This tool is designed to stress test a blockchain node.

- It can create accounts and distribute funds for transaction load testing.
- It allows for flexible load generation through scenario scripting.
- It generates detailed reports.

## Usage

### 1. Generate Accounts

Multiple accounts are needed to generate transaction load.

You can create accounts using the following command

**command** :

```bash
tokamak-trunks account generate
```

**option** :

- `--count-accounts` : number of accounts to generate

**example** :

```bash
tokamak-trunks account generate --count-accounts=20
```

> The private keys for the test accounts are stored in the ~/.tokamak-trunks allowing them to be reused.

### 2. Faucet balance

To generate transaction load, you need to distribute funds to the test accounts.

Please ensure that the distributor account has sufficient funds.

**command** :

```bash
tokamak-trunks account faucet
```

**options** :

- `--rpc-url` : RPC URL
- `--distributor-private-key` : private key for ETH(TON) distribute

**example** :

```bash
tokamak-trunks account faucet \
  --rpc-url=http://localhost:9545 \
  --distributor-private-key=0xABCDEFGHIJKLMN
```

### 3. Load Test

You can create scenarios to conduct load testing.

Scenario files are in YAML format.

**command** :

```bash
tokamak-trunks start
```

**options** :

- `--l1-rpc-url` : L1 RPC URL
- `--l2-rpc-url` : L2 RPC URL
- `--l1-chain-id`: L1 Chain ID
- `--l2-chain-id`: L2 Chain ID
- `--scenario-file-path` : Scenario file path
- `--l2-block-time` : L2 Block Time
- `--output-file-name` : report file name for output

**scenario** :

- `name` : name of test
- `method` : You can select the type of load `call`, `transaction`
- `duration` : This is the duration for which the load will be applied.
- `pace` : Define the attack rate.
  - `linear` : The RPS increases linearly by the magnitude of the slope.
  - `rate` : Define RPS

```yaml
# test scenario

name: "scenario-name"

actions:
  - method: call
    duration: 1m
    pace:
      linear:
        start:
          freq: 10
          per: 1s
        slope: 2
  - method: transaction
    duration: 1m
    pace:
      rate:
        freq: 100
        per: 1s
```

**example** :

```bash
tokamak-trunks start \
  --l1-rpc-url="http://localhost:8545" \
  --l2-rpc-url="http://localhost:9545" \
  --l1-chain-id=900 \
  --l2-chain-id=901 \
  --scenario-file-path="./example_scenario.yaml" \
  --l2-block-time=2
  --output-file-name="example-report"
```

### 4. Report

A report is automatically generated at the end of the test.

```
transaction
Requests      [total, rate, throughput]         2000, 200.11, 166.66
Duration      [total, attack, wait]             11.904s, 9.995s, 1.909s
Latencies     [min, mean, 50, 90, 95, 99, max]  2.051ms, 1.358s, 1.504s, 2.006s, 2.449s, 2.507s, 2.528s
Bytes In      [total, mean]                     205552, 102.78
Bytes Out     [total, mean]                     591996, 296.00
Success       [ratio]                           99.20%
Status Codes  [code:count]                      200:1984  33536:16
Error Set:
err: already known

Transaction report
TPS                           198
Total Confirmed Tx            1983
First Confirmed Block Number  269115
Last Confirmed Block Number   269120
Total Using Gas
  L1Gas    7634644
  BlobGas  0
  L2Gas    41643000
Average Gas Price
  L1Gas    7 Wei (0.000000 Gwei)
  BlobGas  0 Wei (0.000000 Gwei)
  L2Gas    1001262 Wei (0.001001 Gwei)
Total Fee
  L1Fee      53415628 Wei (0.000000 ETH)
  BlobFee    0 Wei (0.000000 ETH)
  L2Fee      41632548825000 Wei (0.000042 ETH)
L2BlockTime  2s
```
