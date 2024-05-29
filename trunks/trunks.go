package trunks

import (
	"fmt"
	"math/big"
	"os"
	"sync"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type Trunks struct {
	wg *sync.WaitGroup

	Scenario *Scenario

	L1RPC string
	L2RPC string

	L1ChainId   *big.Int
	L2ChainId   *big.Int
	L2BlockTime *big.Int

	Accounts *Accounts

	L1StandardBridgeAddress    string
	L2StandardBridgeAddress    string
	L2ToL1MessagePasserAddress string
	BatcherAddress             string
	ProposerAddress            string
	SequencerFeeVaultAddress   string

	outputFileName string
}

func (t *Trunks) Start() error {
	file, err := os.Create(t.Scenario.Name)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, action := range t.Scenario.Actions {
		var metrics vegeta.Metrics
		fmt.Printf("start action %s\n", action.Method)
		file.WriteString(fmt.Sprintf("%s attack\n", action.Method))
		attacker, err := MakeAttacker(&action, t)
		if err != nil {
			return err
		}
		for res := range attacker.Attack() {
			metrics.Add(res)
		}

		metrics.Close()
		reporter := vegeta.NewTextReporter(&metrics)
		reporter.Report(file)
		file.WriteString("\n")
	}
	fmt.Println("eeeeend start")
	return nil
}
