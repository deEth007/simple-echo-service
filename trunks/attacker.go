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

func MakeAttacker(action *Action, t *Trunks) (Attacker, error) {
	duration, err := time.ParseDuration(action.Duration)
	if err != nil {
		return nil, err
	}
	if action.Method == "call" {
		tOption := &TargetOption{
			RPC: t.L2RPC,
		}
		return &CallAttacker{
			Pace:     action.GetPace(),
			Duration: duration,
			Targeter: CallTargeter(tOption),
		}, nil

	}
	if action.Method == "transaction" {
		rpc := t.L2RPC
		chainId := t.L2ChainId
		if action.Bridge == "deposit" {
			rpc = t.L1RPC
			chainId = t.L1ChainId
		}
		client, _ := ethclient.Dial(rpc)
		tOption := &TargetOption{
			RPC: rpc,
			TransactionOption: &TransactionOption{
				Accounts: t.Accounts,
				ChainId:  chainId,
				To:       action.To,
				Client:   client,
			},
		}
		return &TransactionAttacker{
			Client:   client,
			Pace:     action.GetPace(),
			Duration: duration,
			Targeter: TransactionTargeter(tOption),
		}, nil
	}

	return nil, fmt.Errorf("wrong action method")
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
	fmt.Println("transaction attack start")
	attacker := vegeta.NewAttacker()
	sem := semaphore.NewWeighted(3)
	results := make(chan *vegeta.Result)

	go func() {
		wg := new(sync.WaitGroup)
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
				defer wg.Done()
				err := waitTxConfirm(txHash, rr, ta.Client)
				if err != nil {
					return
				}
				results <- rr
			}(res)
			wg.Wait()
		}
	}()
	return results
}

func waitTxConfirm(txHash common.Hash, result *vegeta.Result, client *ethclient.Client) error {
	reporter := reporter.GetTrunksReport()
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
		return common.Hash{}, errors.New("not exist result")
	}

	txHash := common.HexToHash(body["result"].(string))
	return txHash, nil
}
