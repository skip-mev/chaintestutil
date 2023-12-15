package network

import (
	"context"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestSuite is a test suite for tests that initializes a network instance.
type TestSuite struct {
	Network *Network
}

func NewSuite(t *testing.T, cfg network.Config) *TestSuite {
	return &TestSuite{Network: New(t, cfg)}
}

func (s *TestSuite) GetGRPC() (cc *grpc.ClientConn, close func(), err error) {
	// get grpc address
	grpcAddr := s.Network.Validators[0].AppConfig.GRPC.Address

	// create the client
	cc, err = grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	close = func() { cc.Close() }

	return
}

func (s *TestSuite) CreateTxBytes(fees sdk.Coin, gas uint64, msgs []sdk.Msg) ([]byte, error) {
	val := s.Network.Validators[0]

	kr, err := val.ClientCtx.Keyring.KeyByAddress(val.Address)
	if err != nil {
		return nil, err
	}

	txFactory := clienttx.Factory{}.
		WithChainID(val.ClientCtx.ChainID).
		WithKeybase(val.ClientCtx.Keyring).
		WithTxConfig(val.ClientCtx.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).WithFees(fees.String()).
		WithGas(gas).
		WithSequence(1)
	builder, err := txFactory.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}
	err = clienttx.Sign(context.Background(), txFactory, kr.Name, builder, true)
	if err != nil {
		return nil, err
	}

	bz, err := val.ClientCtx.TxConfig.TxEncoder()(builder.GetTx())
	return bz, err
}
