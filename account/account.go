package account

import (
	"bufio"
	"crypto/ecdsa"
	"fmt"
	"os"

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

func GetAccounts() *Accounts {
	accounts := &Accounts{}
	stringPrivateKeys, _ := read(getAccountFilePath())
	for _, key := range stringPrivateKeys {
		privateKey, _ := stringToPrivateKey(key)
		address := getAddress(privateKey)
		newAccount := Account{
			Address: address,
			PrivKey: privateKey,
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

func getAddress(privKey *ecdsa.PrivateKey) common.Address {
	publicKey := privKey.PublicKey
	return crypto.PubkeyToAddress(publicKey)
}

func stringToPrivateKey(key string) (*ecdsa.PrivateKey, error) {
	privateKey, err := crypto.HexToECDSA(key[2:])
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func read(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var data []string
	for scanner.Scan() {
		d := scanner.Text()
		data = append(data, d)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func write(path string, data []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, d := range data {
		_, err := writer.WriteString(d + "\n")
		if err != nil {
			return err
		}
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func getAccountFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return fmt.Sprintf("%s/%s/%s", homeDir, ".tokamak-trunks", "accounts")
}
