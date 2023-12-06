package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/skip-mev/chaintestutil/sample"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/skip-mev/chaintestutil/keeper"
)

func TestTestKeepers_MintToAccount(t *testing.T) {
	sdkCtx, tk, _ := testkeeper.NewTestSetup(t)
	r := sample.Rand()
	ctx := sdk.WrapSDKContext(sdkCtx)
	address := sample.Address(r)
	coins, otherCoins := sample.Coins(r), sample.Coins(r)

	getBalances := func(address string) sdk.Coins {
		res, err := tk.BankKeeper.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
			Address: address,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		return res.Balances
	}

	// should create the account
	tk.MintToAccount(sdkCtx, address, coins)
	require.True(t, getBalances(address).IsEqual(coins))

	// should add the minted coins in the balance
	previousBalance := getBalances(address)
	tk.MintToAccount(sdkCtx, address, otherCoins)
	require.True(t, getBalances(address).IsEqual(previousBalance.Add(otherCoins...)))
}
