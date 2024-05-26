package trunks

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Trunks struct {
	wg *sync.WaitGroup

	L1RPC string
	L2RPC string

	L1ChainId *big.Int
	L2ChainId *big.Int

	TransferAccounts   *Accounts
	DepositAccounts    *Accounts
	WithdrawalAccounts *Accounts

	L1StandardBridgeAddress common.Address
}

func (t *Trunks) CallAttacker(tageter CallTargeterFn) error {
	rate := vegeta.Rate{Freq: 1000, Per: time.Second}
	duration := 10 * time.Second
	attacker := vegeta.NewAttacker()

	tgter := tageter(t)
	results := make(chan *vegeta.Result)
	var metrics vegeta.Metrics

	go func() {
		for res := range results {
			metrics.Add(res)
			metrics.Close()
		}
	}()

	file, err := os.Create("call_results.bin")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := vegeta.NewEncoder(file)

	for res := range attacker.Attack(tgter, rate, duration, "call") {
		results <- res
		if err := encoder.Encode(res); err != nil {
			return err
		}
	}
	close(results)
	t.wg.Done()
	return err
}

func (t *Trunks) TransferAttacker(tageter TrnasactionTageterFn) error {
	rate := vegeta.Rate{Freq: 10, Per: time.Second}
	duration := 10 * time.Second

	client, err := ethclient.Dial(t.L2RPC)
	if err != nil {
		return err
	}
	attacker := vegeta.NewAttacker()

	tgter := tageter(t, client)
	results := make(chan *vegeta.Result)
	var metrics vegeta.Metrics

	go func() {
		for res := range results {
			metrics.Add(res)
			metrics.Close()
		}
	}()

	file, err := os.Create("transfer_results.bin")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := vegeta.NewEncoder(file)

	for res := range attacker.Attack(tgter, rate, duration, "transfer") {
		body := map[string]interface{}{}
		json.Unmarshal(res.Body, &body)
		fmt.Printf("res: %v\n", body)
		txHash := common.HexToHash(body["result"].(string))
		var blockNumber *big.Int
		for {
			receipt, err := client.TransactionReceipt(context.Background(), txHash)
			if err != nil {
				if err == ethereum.NotFound {
					fmt.Println("Transaction is not yet mined")
				} else {
					fmt.Printf("Somthing error: %s\n", err)
				}
			} else {
				blockNumber = receipt.BlockNumber
				res.Code = uint16(receipt.Status)
				fmt.Printf("receipt: %v\n", receipt)
				break
			}
			time.Sleep(time.Second * 2)
		}
		block, err := client.BlockByNumber(context.Background(), blockNumber)
		if err != nil {
			return err
		}
		blockTime := block.Time()
		blockTimeToUnix := time.Unix(int64(blockTime), 0)

		res.Latency = blockTimeToUnix.Sub(res.Timestamp)

		results <- res
		if err := encoder.Encode(res); err != nil {
			return err
		}
	}
	close(results)
	t.wg.Done()
	return nil
}

func (t *Trunks) DepositAttacker() {}

func (t *Trunks) WithdrawalAttacker() {}
