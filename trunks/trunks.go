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
	"golang.org/x/sync/semaphore"

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

	L1StandardBridgeAddress string
	L2StandardBridgeAddress string

	outputFileName string
}

type TxReporter func(*vegeta.Result, *ethclient.Client) *vegeta.Result

func (t *Trunks) Start() {
	opts := &TxOpts{
		TargetRPC:     t.L2RPC,
		TargetChainId: t.L2ChainId,
		Accounts:      t.WithdrawalAccounts,
	}
	pace := vegeta.Rate{Freq: 100, Per: time.Second}
	duration := time.Duration(2 * time.Second)
	t.TransactionAttack(TranferReporter, TransactionTageter, pace, duration, opts)
}

func (t *Trunks) CallAttack(tageter CallTargeterFn) error {
	rate := vegeta.Rate{Freq: 1000, Per: time.Second}
	duration := 2 * time.Second
	attacker := vegeta.NewAttacker()

	tgter := tageter(t)
	results := make(chan *vegeta.Result)

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

func (t *Trunks) TransactionAttack(txReporter TxReporter, tageter TransactionTageterFn, pace vegeta.Pacer, duration time.Duration, opts *TxOpts) {
	client, _ := ethclient.Dial(opts.TargetRPC)
	attacker := vegeta.NewAttacker()
	tgter := tageter(opts)

	file, _ := os.Create(t.outputFileName)
	defer file.Close()
	// encoder := vegeta.NewEncoder(file)
	var metrics vegeta.Metrics
	var txSuccess uint16

	var mu sync.Mutex
	sem := semaphore.NewWeighted(3)

	t.wg.Add(1)
	go func() {
		for res := range attacker.Attack(tgter, pace, duration, "Transaction Attack") {
			sem.Acquire(context.Background(), 1)
			t.wg.Add(1)

			go func(rr *vegeta.Result) {
				defer sem.Release(1)
				r := txReporter(rr, client)
				metrics.Add(r)
				if r.Code == 1 {
					mu.Lock()
					txSuccess++
					mu.Unlock()
				}
				t.wg.Done()
			}(res)
		}
		metrics.Close()
		metrics.Success = float64(txSuccess) / float64(metrics.Requests)
		t.wg.Done()
	}()
	t.wg.Wait()

	fmt.Println("Reporting result...")
	reporter := vegeta.NewTextReporter(&metrics)
	reporter.Report(file)
}

func TranferReporter(result *vegeta.Result, client *ethclient.Client) *vegeta.Result {
	r := result
	body := map[string]interface{}{}
	json.Unmarshal(r.Body, &body)
	_, exist := body["result"]
	if !exist {
		return r
	}
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
			r.Code = uint16(receipt.Status)
			fmt.Printf("receipt: %+v\n", receipt)
			break
		}
		time.Sleep(time.Second * 2)
	}
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		return r
	}
	blockTime := block.Time()
	blockTimeToUnix := time.Unix(int64(blockTime), 0)

	r.Latency = blockTimeToUnix.Sub(r.Timestamp)

	return r
}
