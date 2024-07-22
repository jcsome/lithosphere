# Lithosphere

Lithosphere is the data synchronization hub of Mantle Network, which plays a key role in indexing data between Layer1 and Layer2. In addition to synchronizing L1->L2 deposit and L2->L1 withdraw data, Lithosphere also supports MantleDA and StateRoot data synchronization to meet the needs of the subsequent ecological data support. See [Architecture](./architecture.md) for more details.

## Getting started

### Setup Environment

The `example.env` shows a set of environmental variables that can be used to run the [Lithosphere index service](./architecture.md#lithosphere-service), [Lithosphere API](./architecture.md#lithosphere-api) and Lithosphere exporter.

### Setup Polling Intervals

The Lithosphere polls and processes batches from the L1 and L2 chains on a set interval/size. The default polling interval is 5 seconds for both chains with a default batch header size of 500. The polling frequency can be changed by setting the `L1_POLLING_INTERVAL` and `L2_POLLING_INTERVAL` values in the `.env` file. The batch header size can be changed by setting the `L1_HEADER_BUFFER_SIZE` and `L2_HEADER_BUFFER_SIZE` values in the `.env` file.

> L1 blocks are indexed only if they contain L1 system contract events. Therefore, the `l1_block_headers` table will not contain every L1 block header.

### Run Lithosphere

1. Install docker
2. Run `cp example.env .env`
3. Complete your configuration in `.env`
4. Run `docker compose up` to start the Lithosphere with Optimism Goerli network

### Run Lithosphere CLI

See the flags in `flags.go` for reference of what command line flags to pass to `go run`

### Run Lithosphere in a custom configuration

`docker-compose.dev.yml` is git ignored. Fill in your own docker-compose file here.
