package trunks

import (
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/tokamak-network/tokamak-trunks/nmgr"
	"github.com/tokamak-network/tokamak-trunks/utils"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

type TrunksErvice struct {
	NodeMgr nmgr.NodeManager
	Trunks  *Trunks
}

func Main() cli.ActionFunc {
	return func(cliCtx *cli.Context) error {
		cfg := NewCLIConfig(cliCtx)
		service, err := NewService(cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer service.Stop()
		if err := service.Start(); err != nil {
			return err
		}
		return nil
	}
}

func NewService(cfg *CLIConfig) (*TrunksErvice, error) {
	scenario, err := initScenario(cfg.ScenarioFilePath)
	if err != nil {
		return nil, err
	}

	var accounts *Accounts
	var nodeMgr *nmgr.BaseNodeManager
	if cfg.NodeManagerEnable {
		accounts = initAccounts(scenario.Accounts)
		nodeMgr, err = initBaseNodeManager(cfg, accounts)
		if err != nil {
			return nil, err
		}
	}

	trunks, err := initTrunks(cfg, accounts, scenario)
	if err != nil {
		return nil, err
	}

	return &TrunksErvice{
		NodeMgr: nodeMgr,
		Trunks:  trunks,
	}, nil
}

func initScenario(path string) (*Scenario, error) {
	file, err := os.ReadFile(utils.ConvertToAbsPath(path))
	if err != nil {
		return nil, err
	}

	var scenario Scenario
	err = yaml.Unmarshal(file, &scenario)
	if err != nil {
		return nil, err
	}
	return &scenario, nil
}

func initAccounts(count uint) *Accounts {
	return GenerateAccounts(count)
}

func initBaseNodeManager(cfg *CLIConfig, accounts *Accounts) (*nmgr.BaseNodeManager, error) {
	return nmgr.NewBaseNodeManager(
		nmgr.NewConfig(cfg.NodeMgr, accounts.GetAddresses()...),
	)
}

func initTrunks(cfg *CLIConfig, accounts *Accounts, scenario *Scenario) (*Trunks, error) {

	transferAccounts := &Accounts{
		List: accounts.List[:len(accounts.List)/3+1],
	}
	depositAccounts := &Accounts{
		List: accounts.List[len(accounts.List)/3+1 : (len(accounts.List)*2)/3+2],
	}
	withdrawalAccounts := &Accounts{
		List: accounts.List[(len(accounts.List)*2)/3+2:],
	}

	return &Trunks{
		wg: new(sync.WaitGroup),

		Scenario: scenario,

		L1RPC: cfg.L1RPC,
		L2RPC: cfg.L2RPC,

		L1ChainId:   new(big.Int).SetUint64(cfg.L1ChainId),
		L2ChainId:   new(big.Int).SetUint64(cfg.L2ChainId),
		L2BlockTime: new(big.Int).SetUint64(cfg.L2BlockTime),

		TransferAccounts:   transferAccounts,
		DepositAccounts:    depositAccounts,
		WithdrawalAccounts: withdrawalAccounts,

		L1StandardBridgeAddress:    cfg.L1StandardBrige,
		L2StandardBridgeAddress:    cfg.L2StandardBrige,
		L2ToL1MessagePasserAddress: cfg.L2ToL1MessagePasser,
		BatcherAddress:             cfg.Batcher,
		ProposerAddress:            cfg.Proposer,
		SequencerFeeVaultAddress:   cfg.SequencerFeeVault,

		outputFileName: scenario.Name,
	}, nil
}

func (ts *TrunksErvice) Start() error {
	if ts.NodeMgr != nil {
		err := ts.NodeMgr.Start()
		if err != nil {
			return err
		}
	}
	time.Sleep(10 * time.Second)

	ts.Trunks.Start()

	return nil
}

func (ts *TrunksErvice) Stop() {
	// ts.NodeMgr.Destroy()
	ChainReporter.RecordTPS(big.NewInt(2))
	ChainReporter.PrintReport()
}
