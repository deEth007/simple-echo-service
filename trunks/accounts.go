package trunks

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Account struct {
	Address common.Address
	PrivKey *ecdsa.PrivateKey
}

type Accounts struct {
	List []Account
}

func GenerateAccounts(count uint) *Accounts {
	accounts := &Accounts{}
	for i := uint(0); i < count; i++ {
		newPrivKey, _ := crypto.GenerateKey()
		address := GetAddress(newPrivKey)
		newAccount := Account{
			Address: address,
			PrivKey: newPrivKey,
		}
		accounts.List = append(accounts.List, newAccount)
	}
	return accounts
}

func (a *Accounts) GetAddresses() []common.Address {
	var addresses []common.Address
	for _, account := range a.List {
		addresses = append(addresses, account.Address)
	}
	return addresses
}

func GetAddress(privKey *ecdsa.PrivateKey) common.Address {
	publicKey := privKey.PublicKey
	return crypto.PubkeyToAddress(publicKey)
}
