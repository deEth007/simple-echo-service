package trunks

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var CALL_METHOD []string = []string{
	"eth_blockNumber",
	"eth_chainId",
	"eth_getBalance",
	"eth_gasPrice",
}

func CallTargeter() vegeta.Targeter {
	roundRobin := -1
	return func(tgt *vegeta.Target) error {
		if tgt == nil {
			return vegeta.ErrNilTarget
		}

		tgt.Method = "POST"
		tgt.URL = "http://localhost:9545"
		tgt.Header = map[string][]string{
			"Content-type": []string{"application/json"},
		}

		roundRobin = (roundRobin + 1) % 4
		body := fmt.Sprintf(`{"jsonrpc": "2.0", "method":"%s", "params": [], "id": 0}`, CALL_METHOD[roundRobin])
		tgt.Body = []byte(body)
		return nil
	}
}

var value *big.Int = big.NewInt(1000000000000000000)

func TransferTageter(accounts *Accounts, client *ethclient.Client) vegeta.Targeter {
	roundRobin := -1
	gasPrice, _ := client.SuggestGasPrice(context.Background())
	chainId, _ := client.NetworkID(context.Background())

	return func(tgt *vegeta.Target) error {
		roundRobin = (roundRobin + 1) % len(accounts.List)

		from := accounts.List[roundRobin]
		to := accounts.List[(roundRobin+1)%len(accounts.List)]

		var data []byte
		nonce := NonceCache[from.Address.Hex()]
		tx := types.NewTransaction(nonce, to.Address, value, uint64(21000), gasPrice, data)
		NonceCache[from.Address.Hex()] = nonce + uint64(1)

		signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainId), from.PrivKey)
		rawTxBytes, _ := signedTx.MarshalBinary()

		rawTxHex := hex.EncodeToString(rawTxBytes)
		body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0x%s"],"id":1}`, rawTxHex)

		tgt.Method = "POST"
		tgt.URL = "http://localhost:9545"
		tgt.Header = map[string][]string{
			"Content-type": []string{"application/json"},
		}
		tgt.Body = []byte(body)

		return nil
	}
}
