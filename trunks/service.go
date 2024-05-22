package trunks

import (
	"log"

	"github.com/tokamak-network/tokamak-trunks/nmgr"
	"github.com/urfave/cli/v2"
)

type TrunksErvice struct {
	NodeMgr nmgr.NodeManager
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
	if cfg.NodeManagerEnable {
		var err error
		svc.NodeMgr, err = nmgr.NewBaseNodeManager(
			nmgr.NewConfig(cfg.NodeMgr),
		)
		if err != nil {
			return nil, err
		}
	}
	return &svc, nil
}

func (ts *TrunksErvice) Start() error {
	if ts.NodeMgr != nil {
		err := ts.NodeMgr.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ts *TrunksErvice) Stop() {}
