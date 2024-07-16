package trunks

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/tokamak-network/tokamak-trunks/account"
)

var CALL_METHOD []string = []string{
	"eth_blockNumber",
	"eth_chainId",
	"eth_gasPrice",
}

type TargetOption struct {
	RPC string

	*TransactionOption
}

type TransactionOption struct {
	Accounts *account.Accounts
	ChainId  *big.Int
	To       string
	Client   *ethclient.Client
	Data     []byte
	GasLimit uint64
}

func CallTargeter(opts *TargetOption) vegeta.Targeter {
	roundRobin := -1
	return func(tgt *vegeta.Target) error {
		if tgt == nil {
			return vegeta.ErrNilTarget
		}

		tgt.Method = "POST"
		tgt.URL = opts.RPC
		tgt.Header = map[string][]string{
			"Content-type": {"application/json"},
		}

		roundRobin = (roundRobin + 1) % len(CALL_METHOD)
		body := fmt.Sprintf(
			`{"jsonrpc": "2.0", "method":"%s", "params": [], "id": 0}`,
			CALL_METHOD[roundRobin],
		)
		tgt.Body = []byte(body)
		return nil
	}
}

// 0.000001 ETH
var value *big.Int = big.NewInt(1000000000000)

func TransactionTargeter(opts *TargetOption) vegeta.Targeter {
	roundRobin := -1
	RPC := opts.RPC
	chainId := opts.ChainId
	client := opts.Client
	accounts := opts.Accounts
	data := opts.Data
	gasLimit := opts.GasLimit

	mutex := sync.Mutex{}

	return func(tgt *vegeta.Target) error {
		if tgt == nil {
			return vegeta.ErrNilTarget
		}

		mutex.Lock()
		roundRobin = (roundRobin + 1) % len(accounts.List)
		localRoundRobin := roundRobin
		mutex.Unlock()

		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			return err
		}

		from := accounts.List[localRoundRobin]
		var to common.Address
		if opts.To == "" {
			to = accounts.List[(localRoundRobin+1)%len(accounts.List)].Address
		} else {
			to = common.HexToAddress(opts.To)
		}

		if gasLimit == 0 {
			gasLimit = uint64(300000)
		}
		nonce, err := client.PendingNonceAt(context.Background(), from.Address)
		if err != nil {
			return err
		}

		tx := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, data)

		signedTx, err := types.SignTx(tx, types.NewCancunSigner(chainId), from.PrivKey)
		if err != nil {
			return err
		}

		rawTxBytes, err := signedTx.MarshalBinary()
		if err != nil {
			return err
		}

		rawTxHex := hex.EncodeToString(rawTxBytes)
		body := fmt.Sprintf(
			`{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0x%s"],"id":1}`,
			rawTxHex,
		)

		tgt.Method = "POST"
		tgt.URL = RPC
		tgt.Header = map[string][]string{
			"Content-type": {"application/json"},
		}
		tgt.Body = []byte(body)

		return nil
	}
}
