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
		L1RPC: "http://localhost:8545",
		L2RPC: "http://localhost:9545",

		L1ChainId: big.NewInt(900),
		L2ChainId: big.NewInt(901),

		TransferAccounts:   transferAccounts,
		DepositAccounts:    depositAccounts,
		WithdrawalAccounts: WithdrawalAccounts,

		L1StandardBridgeAddress: common.HexToAddress("0x1c23A6d89F95ef3148BCDA8E242cAb145bf9c0E4"),
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

	ts.Trunks.wg.Add(2)
	go ts.Trunks.CallAttacker(CallTargeter)
	go ts.Trunks.TransferAttacker(TransferTageter)

	ts.Trunks.wg.Wait()
	return nil
}

func (ts *TrunksErvice) Stop() {}
