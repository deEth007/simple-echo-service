package trunks

import (
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tokamak-network/tokamak-trunks/nmgr"
	"github.com/urfave/cli/v2"
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
	var svc TrunksErvice

	transferAccounts := GenerateAccounts(1000)
	depositAccounts := GenerateAccounts(1000)
	WithdrawalAccounts := GenerateAccounts(1000)

	allAddress := []common.Address{}
	allAddress = append(allAddress, transferAccounts.GetAddresses()...)
	allAddress = append(allAddress, depositAccounts.GetAddresses()...)
	allAddress = append(allAddress, WithdrawalAccounts.GetAddresses()...)

	if cfg.NodeManagerEnable {
		var err error
		svc.NodeMgr, err = nmgr.NewBaseNodeManager(
			nmgr.NewConfig(cfg.NodeMgr, allAddress...),
		)
		if err != nil {
			return nil, err
		}
	}

	trunks := &Trunks{
		wg:    new(sync.WaitGroup),
		L1RPC: cfg.L1RPC,
		L2RPC: cfg.L2RPC,

		L1ChainId: new(big.Int).SetUint64(cfg.L1ChainId),
		L2ChainId: new(big.Int).SetUint64(cfg.L2ChainId),

		TransferAccounts:   transferAccounts,
		DepositAccounts:    depositAccounts,
		WithdrawalAccounts: WithdrawalAccounts,

		L1StandardBridgeAddress: cfg.L1StandardBrige,
		L2StandardBridgeAddress: cfg.L2StandardBrige,

		outputFileName: cfg.OutputFileName,
	}

	svc.Trunks = trunks

	return &svc, nil
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

func (ts *TrunksErvice) Stop() {}
