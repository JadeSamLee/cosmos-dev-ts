# HTLC Module Deployment Instructions

This document provides instructions for building and deploying the custom HTLC module to a Cosmos SDK testnet.

## 1. Build the Cosmos SDK Application

Ensure your Cosmos SDK application includes the `x/htlc` module in the app.go and module manager.

Build the application binary:

```bash
make build
```

Or if using Go directly:

```bash
go build -o myapp ./cmd/myapp
```

## 2. Initialize a New Testnet

Create a new chain directory and initialize the node:

```bash
./myapp init mynode --chain-id mychain
```

Generate a key for your validator:

```bash
./myapp keys add validator
```

Add genesis accounts with tokens:

```bash
./myapp add-genesis-account $(./myapp keys show validator -a) 100000000stake
```

Generate a genesis transaction:

```bash
./myapp gentx validator 50000000stake --chain-id mychain
```

Collect genesis transactions:

```bash
./myapp collect-gentxs
```

Validate the genesis file:

```bash
./myapp validate-genesis
```

## 3. Start the Validator Node

Start the node:

```bash
./myapp start
```

## 4. Interact with the HTLC Module

Use CLI commands to create HTLCs, claim, and refund tokens. For example:

```bash
./myapp tx htlc create-htlc [flags]
./myapp tx htlc claim-htlc [flags]
./myapp tx htlc refund-htlc [flags]
```

Refer to the module's CLI documentation for detailed usage.

## 5. Joining an Existing Testnet

If joining an existing testnet, obtain the genesis file and peer addresses, then start the node with:

```bash
./myapp start --p2p.persistent_peers <peer-addresses>
```

## Notes

- Ensure your module is properly registered in the app's module manager.
- Update the app's CLI commands to include the HTLC module commands.
- For IBC tokens, ensure IBC is configured and running.

For further assistance, refer to the Cosmos SDK documentation.
