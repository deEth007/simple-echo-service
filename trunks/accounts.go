package trunks

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
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

func (a *Accounts) InitNonceCache(client *ethclient.Client) error {
	NonceCache = map[string]uint64{}
	for _, account := range a.List {
		nonce, err := client.PendingNonceAt(context.Background(), account.Address)
		if err != nil {
			return err
		}
		NonceCache[account.Address.Hex()] = nonce
	}
	return nil
}

func GetAddress(privKey *ecdsa.PrivateKey) common.Address {
	publicKey := privKey.PublicKey
	return crypto.PubkeyToAddress(publicKey)
}

var NonceCache map[string]uint64
