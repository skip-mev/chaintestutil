package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/skip-mev/chaintestutil/sample"

	testkeeper "github.com/skip-mev/chaintestutil/keeper"
)

func TestTestKeepers_MintToAccount(t *testing.T) {
	ctx, tk, _ := testkeeper.NewTestSetup(t)
	r := sample.Rand()
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
	tk.MintToAccount(ctx, address, coins)
	require.True(t, getBalances(address).Equal(coins))

	// should add the minted coins in the balance
	previousBalance := getBalances(address)
	tk.MintToAccount(ctx, address, otherCoins)
	require.True(t, getBalances(address).Equal(previousBalance.Add(otherCoins...)))
}
