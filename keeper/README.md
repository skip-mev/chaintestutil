# Keeper Testing

The keeper package exposes an easy API for generating a set of "test keepers" and "test message
servers" for performing integration testing on your application at the keeper level.  Typical Cosmos
SDK unit tests may use mocked keepers, or not use a full set that is representative of the full application. 

This setup allows developers to test all keepers in coordination without having to spin up a local network 
or full application.  

The framework initializes and returns keepers for the following modules:
- `x/bank`
- `x/auth`
- `x/staking`
- `x/distribution`
- `x/feegrant`
- `x/mint`

and is easily extendable to add any other keepers that you may with to test upon setup.

A typical testing flow that extends adds the Skip FeeMarket module might look like the following:

```go
import (
	"testing"

	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	testkeeper "github.com/skip-mev/chaintestutil/keeper"
	"github.com/stretchr/testify/require"

	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

// TestKeepers holds all keepers used during keeper tests for all modules
type TestKeepers struct {
	testkeeper.TestKeepers
	FeeMarketKeeper *feemarketkeeper.Keeper
}

// TestMsgServers holds all message servers used during keeper tests for all modules
type TestMsgServers struct {
	testkeeper.TestMsgServers
	FeeMarketMsgServer feemarkettypes.MsgServer
}

var additionalMaccPerms = map[string][]string{
	feemarkettypes.ModuleName:       nil,
	feemarkettypes.FeeCollectorName: {authtypes.Burner},
}

// NewTestSetup returns initialized instances of all the keepers and message servers of the modules
func NewTestSetup(t testing.TB, options ...testkeeper.SetupOption) (sdk.Context, TestKeepers, TestMsgServers) {
	options = append(options, testkeeper.WithAdditionalModuleAccounts(additionalMaccPerms))

	_, tk, tms := testkeeper.NewTestSetup(t, options...)

	// initialize extra keeper
	feeMarketKeeper := FeeMarket(tk.Initializer, tk.AccountKeeper)
	require.NoError(t, tk.Initializer.LoadLatest())

	// initialize msg servers
	feeMarketMsgSrv := feemarketkeeper.NewMsgServer(*feeMarketKeeper)

	ctx := sdk.NewContext(tk.Initializer.StateStore, tmproto.Header{
		Time:   testkeeper.ExampleTimestamp,
		Height: testkeeper.ExampleHeight,
	}, false, log.NewNopLogger())

	err := feeMarketKeeper.SetState(ctx, feemarkettypes.DefaultState())
	require.NoError(t, err)
	err = feeMarketKeeper.SetParams(ctx, feemarkettypes.DefaultParams())
	require.NoError(t, err)

	testKeepers := TestKeepers{
		TestKeepers:     tk,
		FeeMarketKeeper: feeMarketKeeper,
	}

	testMsgServers := TestMsgServers{
		TestMsgServers:     tms,
		FeeMarketMsgServer: feeMarketMsgSrv,
	}

	return ctx, testKeepers, testMsgServers
}

// FeeMarket initializes the fee market module using the testkeepers intializer.
func FeeMarket(
	initializer *testkeeper.Initializer,
	authKeeper authkeeper.AccountKeeper,
) *feemarketkeeper.Keeper {
	storeKey := sdk.NewKVStoreKey(feemarkettypes.StoreKey)
	initializer.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, initializer.DB)

	return feemarketkeeper.NewKeeper(
		initializer.Codec,
		storeKey,
		authKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}
```