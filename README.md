# chaintestutil
Utilities for testing Cosmos SDK chains

## How to use

```shell
go get github.com/skip-mev/chaintestutil
```

## Structure

    .
    ├── encoding                    # Utilities for creating test encoding configurations
    ├── keeper                      # Utilities for creating a set of test keepers for integration tests
    ├── network                     # Utilities for creating a local test network for integration tests
    └── sample                      # Functions for generating randomized Cosmos SDK types such as Coins, sdk.Ints, etc.
