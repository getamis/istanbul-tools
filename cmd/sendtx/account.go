package main

import (
	"crypto/ecdsa"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Account struct {
	priKey *ecdsa.PrivateKey
}

func GenerateAccount() (*Account, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return &Account{
		priKey: key,
	}, nil
}

// private has no '0x' prefix
func AccountFromECDSA(private string) (*Account, error) {
	key, err := crypto.HexToECDSA(private)
	if err != nil {
		return nil, err
	}
	return &Account{
		priKey: key,
	}, nil
}

func (a *Account) PrivateKey() *ecdsa.PrivateKey {
	// hex.EncodeToString(crypto.FromECDSA(a.priKey))
	return a.priKey
}

func (a *Account) PublicKey() *ecdsa.PublicKey {
	// hex.EncodeToString(crypto.FromECDSAPub(&a.priKey.PublicKey))
	return &a.priKey.PublicKey
}

func (a *Account) Address() common.Address {
	return crypto.PubkeyToAddress(a.priKey.PublicKey)
}

func (a *Account) String() string {
	return hex.EncodeToString(crypto.FromECDSA(a.priKey))
}
