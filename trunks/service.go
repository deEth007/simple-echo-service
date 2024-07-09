package trunks

import (
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/tokamak-network/tokamak-trunks/nmgr"
	"github.com/tokamak-network/tokamak-trunks/reporter"
	"github.com/tokamak-network/tokamak-trunks/utils"
)

type TrunksErvice struct {
	Trunks *Trunks
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
	initReporter(cfg)

	scenario, err := initScenario(cfg.ScenarioFilePath)
	if err != nil {
		return nil, err
	}

	trunks, err := initTrunks(cfg, accounts, scenario)
	if err != nil {
		return nil, err
	}

	return &TrunksErvice{
		Trunks: trunks,
	}, nil
}

func initReporter(cfg *CLIConfig) {
	reporter.InitReporter(
		reporter.NewConfig(cfg.Reporter),
	)
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
	return &Trunks{
		wg: new(sync.WaitGroup),

		Scenario: scenario,

		L1RPC: cfg.L1RPC,
		L2RPC: cfg.L2RPC,

		L1ChainId:   new(big.Int).SetUint64(cfg.L1ChainId),
		L2ChainId:   new(big.Int).SetUint64(cfg.L2ChainId),
		L2BlockTime: new(big.Int).SetUint64(cfg.L2BlockTime),

		Accounts: accounts,
	}, nil
}

func (ts *TrunksErvice) Start() error {
	ts.Trunks.Start()
	return nil
}

func (ts *TrunksErvice) Stop() {
	tReport := reporter.GetTrunksReport()
	tReport.RecordTPS()
	reporter.GetReportManager().Report(reporter.TrunksReporter(), "test")
	reporter.GetReportManager().Close()
}
