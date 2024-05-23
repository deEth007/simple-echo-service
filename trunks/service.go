package trunks

import (
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tokamak-network/tokamak-trunks/nmgr"
	"github.com/urfave/cli/v2"
)

// this will be deleted
var testAccount = []common.Address{
	common.HexToAddress("0x3cBb18D55249d2F3e3e99385d12Be29dFfAeE79a"),
	common.HexToAddress("0x3cBb18D55249d2F3e3e99385d12Be29dFfAeE79a"),
	common.HexToAddress("0x7544A2b1B60b10c398442B8de947e421007c72d3"),
	common.HexToAddress("0xa9327f67F1dD7f00b335f9CfB64BEc70D36Ed0cE"),
	common.HexToAddress("0xA3401EdF55bFFa2832b18EAd97f84E3e94AE4EB1"),
	common.HexToAddress("0x9E628CaAd7A6dD3ce48E78812241B41BdbeF6244"),
	common.HexToAddress("0x92546dE8ebC70E236C8F27f0B0aE254309B84332"),
	common.HexToAddress("0xe9b04f9a46d1C73Cf501966B393935DAB639254b"),
	common.HexToAddress("0xa5Ba0109De7BE1a3E89D8eF0430104717D7879bf"),
	common.HexToAddress("0xA961B0D6dcE82dB098cF70A42A14Add3eE3Db2D5"),
}

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
		// svc.NodeMgr.Faucet(testAccount)
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
