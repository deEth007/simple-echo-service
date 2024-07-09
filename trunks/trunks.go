package trunks

import (
	"fmt"
	"math/big"
	"sync"

	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/tokamak-network/tokamak-trunks/account"
	"github.com/tokamak-network/tokamak-trunks/reporter"
)

type Trunks struct {
	wg *sync.WaitGroup

	Scenario *Scenario

	L1RPC string
	L2RPC string

	L1ChainId   *big.Int
	L2ChainId   *big.Int
	L2BlockTime *big.Int

	Accounts *account.Accounts
}

func (t *Trunks) Start() error {
	for _, action := range t.Scenario.Actions {
		var metrics vegeta.Metrics
		fmt.Printf("start action %s\n", action.Method)
		attacker, err := MakeAttacker(&action, t)
		if err != nil {
			return err
		}
		for res := range attacker.Attack() {
			metrics.Add(res)
		}

		metrics.Close()
		vReporter := vegeta.NewTextReporter(&metrics)
		reporter.GetReportManager().Report(vReporter, action.Method)
	}
	return nil
}
