# Network Testing

The network package implements and exposes a fully operational in-process CometBFT
test network that consists of at least one or potentially many validators. This
test network can be used primarily for integration tests or unit test suites.

The test network utilizes SimApp as the ABCI application and uses all the modules
defined in the Cosmos SDK. An in-process test network can be configured with any
number of validators as well as account funds and even custom genesis state.

When creating a test network, a series of Validator objects are returned. Each
Validator object has useful information such as their address and public key. A
Validator will also provide its RPC, P2P, and API addresses that can be useful
for integration testing. In addition, a CometBFT local RPC client is also provided
which can be handy for making direct RPC calls to CometBFT.

Note, due to limitations in concurrency and the design of the RPC layer in
CometBFT, only the first Validator object will have an RPC and API client
exposed. Due to this exact same limitation, only a single test network can exist
at a time. A caller must be certain it calls Cleanup after it no longer needs
the network.

This package is extended from the Cosmos-SDK network testutil [package](https://github.com/cosmos/cosmos-sdk/tree/main/testutil/network).
This package creates a simpler API for setting up your custom application for network testing.

A typical testing flow that extends the bank genesis state might look like the following:
```go
    import (
        "math/rand"

        tmdb "github.com/cometbft/cometbft-db"
        tmrand "github.com/cometbft/cometbft/libs/rand"
        "github.com/cosmos/cosmos-sdk/baseapp"
        servertypes "github.com/cosmos/cosmos-sdk/server/types"
        banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
        pruningtypes "github.com/cosmos/cosmos-sdk/store/pruning/types"
        simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
        "github.com/cosmos/gogoproto/proto"

        "github.com/skip-mev/chaintestutil/network"
        "github.com/skip-mev/chaintestutil/sample"
        "github.com/stretchr/testify/require"   
        "github.com/stretchr/testify/suite"

        "github.com/test-repo/app" // example test repository. Replace with your own 		
    )

    var (
        chainID = "chain-" + tmrand.NewRand().Str(6)

        DefaultAppConstructor = func(val network.ValidatorI) servertypes.Application {
            return app.New(
                val.GetCtx().Logger,
                tmdb.NewMemDB(),
                nil,
                true,
                simtestutil.EmptyAppOptions{},
                baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
                baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
                baseapp.SetChainID(chainID),
            )
	    }
    )

    // NetworkTestSuite is a test suite for query tests that initializes a network instance. 
    type NetworkTestSuite struct {
        suite.Suite
		
        Network        *network.Network
        BankState banktypes.GenesisState
    }

    // SetupSuite setups the local network with a genesis state.
    func (nts *NetworkTestSuite) SetupSuite() {
            var (
                r   = sample.Rand()
                cfg = network.NewConfig(DefaultAppConstructor, app.ModuleBasics, chainID)
            )

            updateGenesisConfigState := func(moduleName string, moduleState proto.Message) {
            buf, err := cfg.Codec.MarshalJSON(moduleState)
            require.NoError(nts.T(), err)
            cfg.GenesisState[moduleName] = buf
        }

        // initialize new bank state
        require.NoError(nts.T(), cfg.Codec.UnmarshalJSON(cfg.GenesisState[banktypes.ModuleName], &nts.BankState))
        nts.BankState = populateBankState(r, nts.BankState)
        updateGenesisConfigState(banktypes.ModuleName, &nts.BankState)

        nts.Network = network.New(nts.T(), cfg)
    }

    func populateBankState(_ *rand.Rand, bankState banktypes.GenesisState) banktypes.GenesisState {
        // intercept and populate the state randomly if desired
		// ...
        return bankState
    }

	func (nts *NetworkTestSuite) TestQueryBalancesRequestHandlerFn() {
        val := s.network.Validators[0]
        baseURL := val.APIAddress

        // Use baseURL to make API HTTP requests or use val.RPCClient to make direct
        // CometBFT RPC calls.
        // ...
    }

    func TestNetworkTestSuite(t *testing.T) {
        suite.Run(t, new(NetworkTestSuite))
    }
```
