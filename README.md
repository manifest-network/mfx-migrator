[![build](https://img.shields.io/circleci/build/github/manifest-network/mfx-migrator/main)](https://app.circleci.com/pipelines/github/manifest-network/mfx-migrator)
[![coverage](https://img.shields.io/codecov/c/github/manifest-network/mfx-migrator)](https://app.codecov.io/gh/manifest-network/mfx-migrator)
[![Go Report Card](https://goreportcard.com/badge/github.com/manifest-network/mfx-migrator)](https://goreportcard.com/report/github.com/manifest-network/mfx-migrator)


**mfx-migrator** is a centralized daemon responsible for migrating data from the old MANY chain to the new MANIFEST chain. 
The daemon analyses the MANY chain for new migration-type transactions, processed them, and triggers token transaction on the MANIFEST chain.

This software is not for external use.

The complete migration flow is as follows:
1. The user performs a migration transaction on the MANY chain using Alberto's Token Migration Portal.
2. The transaction is processed by the MANY chain.
3. The `Talib` block explorer picks up the transaction and stores it in its database.
4. The `mfx-migrator` daemon claim new work items from the `Talib` database.
5. The `mfx-migrator` daemon processes the work item and triggers a token transaction on the MANIFEST chain.
6. The transaction is processed by the MANIFEST chain.
7. The `mfx-migrator` daemon updates the work item status in the `Talib` database.

# Requirements

- Go programming language, version 1.23 or higher
- GNU Make
- Bash
- (Optional) Docker (for running the E2E tests)

# How to use

This section describes how to use the `mfx-migrator` software.

Global flags:
- `-l, --logLevel string` - Set the log level. Possible values are `debug`, `info`, `warn`, and `error`. Default is `info`.
- `--neighborghood uint` - The neighborhood ID to use. Default is 0.
- `--password string` - The password to use for the remote database auth. Default is an empty string.
- `--url string` - The root URL of the remote database API. Default is an empty string.
- `--username string` - The username to use for the remote database auth. Default is an empty string.

## Claim a work item

To claim a work item, run the following command:

```bash
mfx-migrator claim
```

Optional flags:
- `--force` - Force the claim of a work item regardless of its status.
- `--uuid string` - Claim a specific work item by UUID.

This command claims a work item from the remote database and store it in a file in the current directory. 
The file is named `[UUID].json`, where `[UUID]` is the UUID of the work item.
The work item will be locked to prevent other workers from claiming it.

## Migrate a work item

To migrate a claimed work item, run the following command:

```bash
mfx-migrator migrate [UUID]
```
where `[UUID]` is the UUID of the work item.

Flags:
- `--address-prefix string` - Address prefix of the MANIFEST chain. Default is `manifest`.
- `--bank-address string` - The address of the bank account to use for the token transaction on the MANIFEST chain. Default is `bank`.
- `--binary` - The name of the chain binary used to perform the migration. The binary must be in `$PATH`. Default is `manifestd`
- `--chain-home` - The root directory of the chain configuration. Default is an empty string.
- `--chain-id string` - The chain ID of the MANIFEST chain. Default is `manifest-1`.
- `--fee-granter` - The address of the fee granter account to use for the token transaction on the MANIFEST chain. Default is an empty string.
- `--gas-adjustment` - Gas adjustment to use for transactions.
- `--gas-denom` - Denomination of the gas fee.
- `--gas-price` - Minimum gas price to use for transactions
- `--keyring-backend string` - The keyring backend to use. Default is `test`.
- `--node-address` - The RPC endpoint of the MANIFEST chain. Default is `http://localhost:26657`.
- `--uuid string` - The UUID of the work item to migrate. Default is an empty string.
- `--wait-for-block-timeout` - Number of seconds spent waiting for the block to be committed.
- `--wait-for-tx-timeout` - Number of seconds spent waiting for the transaction to be included in a block.

This command triggers a token transaction on the MANIFEST chain and updates the work item status in the remote database.

## Verify a work item

To verify a work item, run the following command:

```bash
mfx-migrator verify [UUID]
```
where `[UUID]` is the UUID of the work item.

Flags:
- `--uuid string` - The UUID of the work item to verify. Default is an empty string.

This command verifies the status of the work item in the remote database.

# Developers

Use the provided `Makefile` to execute common operations

```shell
help                           Display this help screen
build                          Build the project
clean                          Clean the project
install                        Install the project
lint                           Run linter (golangci-lint)
format                         Run formatter (goimports)
govulncheck                    Run govulncheck
vet                            Run go vet
coverage                       Run coverage report
test                           Run tests
```