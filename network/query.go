package network

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/skip-mev/chaintestutil/account"
)

func (s *TestSuite) AccountI(acc account.Account) (sdk.AccountI, error) {
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

func (s *TestSuite) Balances(acc account.Account) (sdk.Coins, error) {
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

func (s *TestSuite) AllValidators() ([]stakingtypes.Validator, error) {
	cc, closeFn, err := s.GetGRPC()
	if err != nil {
		return nil, err
	}
	defer closeFn()

	stakingClient := stakingtypes.NewQueryClient(cc)

	resp, err := stakingClient.Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Status:     "",
		Pagination: nil,
	})
	if err != nil {
		return nil, err
	}

	return resp.Validators, nil
}

func (s *TestSuite) ValidatorDelegations(valAddr string) ([]stakingtypes.DelegationResponse, error) {
	cc, closeFn, err := s.GetGRPC()
	if err != nil {
		return nil, err
	}
	defer closeFn()

	stakingClient := stakingtypes.NewQueryClient(cc)

	resp, err := stakingClient.ValidatorDelegations(context.Background(), &stakingtypes.QueryValidatorDelegationsRequest{
		ValidatorAddr: valAddr,
		Pagination:    nil,
	})
	if err != nil {
		return nil, err
	}

	return resp.DelegationResponses, nil
}

func (s *TestSuite) ValidatorDistributionInfo(valAddr string) (*distrtypes.QueryValidatorDistributionInfoResponse, error) {
	cc, closeFn, err := s.GetGRPC()
	if err != nil {
		return nil, err
	}
	defer closeFn()

	distrClient := distrtypes.NewQueryClient(cc)

	return distrClient.ValidatorDistributionInfo(context.Background(), &distrtypes.QueryValidatorDistributionInfoRequest{
		ValidatorAddress: valAddr,
	})
}
