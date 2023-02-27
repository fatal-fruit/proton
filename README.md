# Proton
**Proton** is a sample blockchain app built using Cosmos SDK `v0.47.0-rc2`, IBC `v7.0.0-rc0`, and Tendermint `v0.37.0-rc2`

## Get started

### Prerequisites
- Go `1.19`

### Build/Install
```shell
// Build proton binary
make build

// Or install locally
make install
```

## Tests

### Run Tests
```shell
make test-unit

# TODO: Write e2e tests
make test-e2e
```

### Liveness
Initialize and bring up a local proton chain to test liveness.
```shell
make localnet
```

## Releasing

Create a new tagged release
```shell
git pull --tags --dry-run
git pull --tags
git tag v0.1 -m'Release v0.1'
git push --tags --dry-run
git push origin v0.1
```
