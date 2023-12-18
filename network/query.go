package network

import (
	"context"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/skip-mev/chaintestutil/account"
)

func (s *TestSuite) GetAccountI(acc account.Account) (sdk.AccountI, error) {
	cc, closeFn, err := s.GetGRPC()
	if err != nil {
		return nil, err
	}
	defer closeFn()

	authClient := authtypes.NewQueryClient(cc)

	resp, err := authClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: acc.Address().String()})
	if err != nil {
		return nil, err
	}

	var accI sdk.AccountI
	if err := cdc.UnpackAny(resp.Account, &accI); err != nil {
		return nil, err
	}

	return accI, err
}

func (s *TestSuite) GetBalances(acc account.Account) (sdk.Coins, error) {
	cc, closeFn, err := s.GetGRPC()
	if err != nil {
		return nil, err
	}
	defer closeFn()

	bankClient := banktypes.NewQueryClient(cc)

	resp, err := bankClient.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{
		Address:      acc.Address().String(),
		Pagination:   nil,
		ResolveDenom: false,
	})
	if err != nil {
		return nil, err
	}

	return resp.Balances, nil
}
