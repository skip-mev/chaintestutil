package network

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	cmthttp "github.com/cometbft/cometbft/rpc/client/http"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/skip-mev/chaintestutil/account"
)

var authCodec *codec.ProtoCodec

func init() {
	ir := codectypes.NewInterfaceRegistry()

	authtypes.RegisterInterfaces(ir)
	cryptocodec.RegisterInterfaces(ir)
	authCodec = codec.NewProtoCodec(ir)
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
	if err := authCodec.UnpackAny(resp.Account, &accI); err != nil {
		return nil, err
	}

	return accI, err
}

func (s *TestSuite) GetCometClient() (*cmthttp.HTTP, error) {
	return cmthttp.New(s.Network.Validators[0].RPCAddress, "/websocket")
}

// CreateTx creates and signs a transaction, from the given messages
func (s *TestSuite) CreateTx(ctx context.Context, acc account.Account, gasLimit, fee, timeoutHeight uint64, chainID string, msgs ...sdk.Msg) ([]byte, error) {
	accI, err := s.GetAccountI(acc)
	if err != nil {
		return nil, err
	}

	txConfig := s.Network.Validators[0].ClientCtx.TxConfig

	txFactory := clienttx.Factory{}.
		WithChainID(chainID).
		WithTxConfig(txConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithSequence(accI.GetSequence())
	builder, err := txFactory.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	if err := builder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	// set params
	builder.SetGasLimit(gasLimit)
	builder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(int64(fee)))))
	builder.SetTimeoutHeight(timeoutHeight)

	sigV2 := signing.SignatureV2{
		PubKey: acc.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  txFactory.SignMode(),
			Signature: nil,
		},
		Sequence: accI.GetSequence(),
	}

	if err := builder.SetSignatures(sigV2); err != nil {
		return nil, err
	}

	// now actually sign
	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: accI.GetAccountNumber(),
		Sequence:      accI.GetSequence(),
		PubKey:        acc.PubKey(),
	}

	sigV2, err = clienttx.SignWithPrivKey(
		ctx, signing.SignMode(txConfig.SignModeHandler().DefaultMode()), signerData,
		builder, acc.PrivKey(), txConfig, accI.GetSequence(),
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
