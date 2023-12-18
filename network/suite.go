package network

import (
	"context"
	"github.com/skip-mev/chaintestutil/encoding"
	"testing"

	cmthttp "github.com/cometbft/cometbft/rpc/client/http"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/skip-mev/chaintestutil/account"
)

var cdc *codec.ProtoCodec

func init() {
	cfg := encoding.MakeTestEncodingConfig()
	cdc = codec.NewProtoCodec(cfg.InterfaceRegistry)
}

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

// CreateValidatorTxBytes creates tx bytes using the first validators keyring.
func (s *TestSuite) CreateValidatorTxBytes(fees sdk.Coin, gas uint64, msgs []sdk.Msg) ([]byte, error) {
	val := s.Network.Validators[0]

	kr, err := val.ClientCtx.Keyring.KeyByAddress(val.Address)
	if err != nil {
		return nil, err
	}

	txFactory := clienttx.Factory{}.
		WithChainID(s.Network.Config.ChainID).
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

func (s *TestSuite) GetCometClient() (*cmthttp.HTTP, error) {
	return cmthttp.New(s.Network.Validators[0].RPCAddress, "/websocket")
}

// TxGenInfo contains common info for generating transactions for tests.
type TxGenInfo struct {
	Account       account.Account
	GasLimit      uint64
	TimeoutHeight uint64
	Fee           sdk.Coins
	// OverrideSequence will manually set the account sequence for signing using Sequence.
	OverrideSequence bool
	// Sequence is the account sequence to be used if OverrideSequence is true.
	Sequence uint64
}

// CreateTxBytes creates and signs a transaction, from the given messages.
func (s *TestSuite) CreateTxBytes(ctx context.Context, txGen TxGenInfo, msgs ...sdk.Msg) ([]byte, error) {
	accI, err := s.GetAccountI(txGen.Account)
	if err != nil {
		return nil, err
	}

	txConfig := s.Network.Validators[0].ClientCtx.TxConfig
	sequence := accI.GetSequence()
	if txGen.OverrideSequence {
		sequence = txGen.Sequence
	}

	txFactory := clienttx.Factory{}.
		WithChainID(s.Network.Config.ChainID).
		WithTxConfig(txConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithSequence(sequence)
	builder, err := txFactory.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	if err := builder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	// set params
	builder.SetGasLimit(txGen.GasLimit)
	builder.SetFeeAmount(txGen.Fee)
	builder.SetTimeoutHeight(txGen.TimeoutHeight)

	sigV2 := signing.SignatureV2{
		PubKey: txGen.Account.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  txFactory.SignMode(),
			Signature: nil,
		},
		Sequence: sequence,
	}

	if err := builder.SetSignatures(sigV2); err != nil {
		return nil, err
	}

	// now actually sign
	signerData := authsigning.SignerData{
		ChainID:       s.Network.Config.ChainID,
		AccountNumber: accI.GetAccountNumber(),
		Sequence:      sequence,
		PubKey:        txGen.Account.PubKey(),
	}

	sigV2, err = clienttx.SignWithPrivKey(
		ctx, signing.SignMode(txConfig.SignModeHandler().DefaultMode()), signerData,
		builder, txGen.Account.PrivKey(), txConfig, sequence,
	)
	if err != nil {
		return nil, err
	}

	if err := builder.SetSignatures(sigV2); err != nil {
		return nil, err
	}

	// return tx
	return txConfig.TxEncoder()(builder.GetTx())
}
