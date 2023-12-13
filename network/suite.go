package network

import (
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
