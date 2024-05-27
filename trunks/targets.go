package trunks

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var CALL_METHOD []string = []string{
	"eth_blockNumber",
	"eth_chainId",
	"eth_getBalance",
	"eth_gasPrice",
}

type CallTargeterFn func(*Trunks) vegeta.Targeter
type TransactionTageterFn func(*TxOpts) vegeta.Targeter

func CallTargeter(trunks *Trunks) vegeta.Targeter {
	roundRobin := -1
	return func(tgt *vegeta.Target) error {
		if tgt == nil {
			return vegeta.ErrNilTarget
		}

		tgt.Method = "POST"
		tgt.URL = trunks.L2RPC
		tgt.Header = map[string][]string{
			"Content-type": []string{"application/json"},
		}

		roundRobin = (roundRobin + 1) % len(CALL_METHOD)
		body := fmt.Sprintf(`{"jsonrpc": "2.0", "method":"%s", "params": [], "id": 0}`, CALL_METHOD[roundRobin])
		tgt.Body = []byte(body)
		return nil
	}
}

var value *big.Int = big.NewInt(100000000000000000)

type TxOpts struct {
	TargetRPC     string
	TargetChainId *big.Int
	Accounts      *Accounts
	To            string
}

func TransactionTageter(opts *TxOpts) vegeta.Targeter {
	roundRobin := -1
	RPC := opts.TargetRPC
	chainId := opts.TargetChainId
	client, _ := ethclient.Dial(opts.TargetRPC)
	gasPrice, _ := client.SuggestGasPrice(context.Background())
	accounts := opts.Accounts

	return func(tgt *vegeta.Target) error {
		if tgt == nil {
			return vegeta.ErrNilTarget
		}
		roundRobin = (roundRobin + 1) % len(accounts.List)

		from := accounts.List[roundRobin]
		var to common.Address
		if opts.To == "" {
			to = accounts.List[(roundRobin+1)%len(accounts.List)].Address
		} else {
			to = common.HexToAddress(opts.To)
		}
		fmt.Printf("%s\n", opts.To)
		fmt.Printf("trasnfer from: %s, to: %s\n", from.Address.Hex(), to)

		var data []byte
		transferGas := uint64(3000000)
		nonce, _ := client.PendingNonceAt(context.Background(), from.Address)
		tx := types.NewTransaction(nonce, to, value, transferGas, gasPrice, data)

		signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainId), from.PrivKey)
		rawTxBytes, _ := signedTx.MarshalBinary()

		rawTxHex := hex.EncodeToString(rawTxBytes)
		body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0x%s"],"id":1}`, rawTxHex)

		tgt.Method = "POST"
		tgt.URL = RPC
		tgt.Header = map[string][]string{
			"Content-type": []string{"application/json"},
		}
		tgt.Body = []byte(body)

		return nil
	}
}

// func DepositTageter(trunks *Trunks, client *ethclient.Client) vegeta.Targeter {
// 	roundRobin := -1
// 	RPC := trunks.L1RPC
// 	chainId := trunks.L1ChainId
// 	accounts := trunks.DepositAccounts
// 	l1StandardBridgeAddress := trunks.L1StandardBridgeAddress
// 	transactor, _ := bindings.NewL1StandardBridgeTransactor(l1StandardBridgeAddress, client)

// 	return func(tgt *vegeta.Target) error {
// 		test, _ := ethclient.Dial(RPC)
// 		roundRobin = (roundRobin + 1) % len(accounts.List)

// 		sender := accounts.List[roundRobin]
// 		balance, _ := test.BalanceAt(context.Background(), sender.Address, nil)
// 		nonce, _ := test.PendingNonceAt(context.Background(), sender.Address)
// 		fmt.Printf("sender: %s balance: %d\n nonce: %d\n", sender.Address.Hex(), balance, nonce)

// 		opts, _ := bind.NewKeyedTransactorWithChainID(sender.PrivKey, chainId)
// 		opts.Value = value
// 		opts.Nonce = new(big.Int).SetUint64(nonce)
// 		opts.NoSend = true

// 		tx, _ := transactor.DepositETH(opts, uint32(21000), []byte{})
// 		rawTxBytes, _ := tx.MarshalBinary()
// 		rawTxHex := hex.EncodeToString(rawTxBytes)
// 		body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0x%s"],"id":1}`, rawTxHex)

// 		tgt.Method = "POST"
// 		tgt.URL = RPC
// 		tgt.Header = map[string][]string{
// 			"Content-type": []string{"application/json"},
// 		}
// 		tgt.Body = []byte(body)

// 		return nil
// 	}
// }

// func WithrawTageter() vegeta.Targeter {
// 	return func(t *vegeta.Target) error {
// 		return nil
// 	}
// }
