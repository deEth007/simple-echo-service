package trunks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/tokamak-network/tokamak-trunks/reporter"
	vegeta "github.com/tsenart/vegeta/v12/lib"
	"golang.org/x/sync/semaphore"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Attacker interface {
	Attack() <-chan *vegeta.Result
}

type CallAttacker struct {
	Pace     vegeta.Pacer
	Duration time.Duration
	Targeter vegeta.Targeter
}
type TransactionAttacker struct {
	Client   *ethclient.Client
	Pace     vegeta.Pacer
	Duration time.Duration
	Targeter vegeta.Targeter
}

func (ca *CallAttacker) Attack() <-chan *vegeta.Result {
	fmt.Println("call attack start")
	attacker := vegeta.NewAttacker()
	results := make(chan *vegeta.Result)

	go func() {
		defer close(results)
		for res := range attacker.Attack(ca.Targeter, ca.Pace, ca.Duration, "call") {
			results <- res
		}
	}()
	return results
}

func (ta *TransactionAttacker) Attack() <-chan *vegeta.Result {
	wg := new(sync.WaitGroup)
	fmt.Println("transaction attack start")
	attacker := vegeta.NewAttacker()
	results := make(chan *vegeta.Result)

	go func() {
		sem := semaphore.NewWeighted(3)
		defer close(results)
		for res := range attacker.Attack(ta.Targeter, ta.Pace, ta.Duration, "transaction attack") {
			txHash, err := txHashFromResult(res)
			if err != nil {
				res.Error = err.Error()
				continue
			}
			sem.Acquire(context.Background(), 1)
			wg.Add(1)
			go func(rr *vegeta.Result) {
				defer sem.Release(1)
				err := waitTxConfirm(txHash, rr, ta.Client)
				if err != nil {
					return
				}
				results <- rr
				wg.Done()
			}(res)
		}
	}()
	return results
}

func waitTxConfirm(txHash common.Hash, result *vegeta.Result, client *ethclient.Client) error {
	reporter := reporter.Get()
	var blockNumber *big.Int

	for {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			blockNumber = receipt.BlockNumber
			reporter.RecordStartToLastBlock(receipt)
			reporter.RecordL1GasUsed(receipt)
			reporter.RecordL1GasFee(receipt)
			reporter.RecordL2GasUsed(receipt)
			reporter.RecordL2GasFee(receipt)
			reporter.RecordConfirmRequest()
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

	result.Latency = blockTimeToUnix.Sub(result.Timestamp)

	return nil
}

func txHashFromResult(r *vegeta.Result) (common.Hash, error) {
	body := map[string]interface{}{}
	json.Unmarshal(r.Body, &body)

	if _, exist := body["error"]; exist {
		fmt.Println("error exist")
		return common.Hash{}, errors.New("not exist result")
	}

	txHash := common.HexToHash(body["result"].(string))
	return txHash, nil
}
