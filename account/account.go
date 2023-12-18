package account

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Account is an abstraction around a cosmos-sdk private-key (account)
type Account struct {
	pk cryptotypes.PrivKey
}

// NewAccount returns a new account, with a randomly generated private-key.
func NewAccount() *Account {
	return &Account{
		pk: secp256k1.GenPrivKey(),
	}
}

// Address returns the address of the account.
func (a *Account) Address() sdk.AccAddress {
	return sdk.AccAddress(a.pk.PubKey().Address())
}

// PubKey returns the public-key of the account.
func (a *Account) PubKey() cryptotypes.PubKey {
	return a.pk.PubKey()
}

// PrivKey returns the private-key of the account.
func (a *Account) PrivKey() cryptotypes.PrivKey {
	return a.pk
}

func (acc *Account) Equals(acc2 Account) bool {
	return acc.Address().Equals(acc2.Address())
}
