// Package network allows to programmatically spin up a local network for CLI tests
package network

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/depinject"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
)

type TestApp interface {
	runtime.AppI
	servertypes.Application
}

type (
	Network        = network.Network
	Config         = network.Config
	ValidatorI     = network.ValidatorI
	AppConstructor func(val ValidatorI) TestApp
)

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t *testing.T, cfg network.Config) *network.Network {
	net, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	t.Cleanup(net.Cleanup)
	return net
}

// NewConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig
func NewConfig(appConfig depinject.Config) network.Config {
	cfg, err := network.DefaultConfigWithAppConfig(appConfig)
	if err != nil {
		panic(err)
	}

	cfg.AccountRetriever = authtypes.AccountRetriever{}
	cfg.TimeoutCommit = 2 * time.Second
	cfg.NumValidators = 1
	cfg.BondDenom = sdk.DefaultBondDenom
	cfg.MinGasPrices = fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom)
	cfg.AccountTokens = sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction)
	cfg.StakingTokens = sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction)
	cfg.BondedTokens = sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	cfg.PruningStrategy = pruningtypes.PruningOptionNothing
	cfg.CleanupDir = true
	cfg.SigningAlgo = string(hd.Secp256k1Type)
	cfg.KeyringOptions = []keyring.Option{}

	return cfg
}
