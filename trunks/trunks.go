package trunks

import (
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Trunks struct {
	wg *sync.WaitGroup

	Scenario *Scenario

	L1RPC string
	L2RPC string

	L1ChainId   *big.Int
	L2ChainId   *big.Int
	L2BlockTime *big.Int

	TransferAccounts   *Accounts
	DepositAccounts    *Accounts
	WithdrawalAccounts *Accounts

	L1StandardBridgeAddress    string
	L2StandardBridgeAddress    string
	L2ToL1MessagePasserAddress string
	BatcherAddress             string
	ProposerAddress            string
	SequencerFeeVaultAddress   string

	outputFileName string
}

func (t *Trunks) Start() error {
	var metrics vegeta.Metrics
	file, err := os.Create("call_results")
	if err != nil {
		return err
	}
	defer file.Close()

	for _, action := range t.Scenario.Actions {
		fmt.Printf("start action %s\n", action.Method)
		attacker, err := MakeAttacker(&action, t)
		if err != nil {
			return err
		}
		for res := range attacker.Attack() {
			metrics.Add(res)
		}
	}
	metrics.Close()

	reporter := vegeta.NewTextReporter(&metrics)
	reporter.Report(file)

	return nil
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
				Accounts: t.TransferAccounts,
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
