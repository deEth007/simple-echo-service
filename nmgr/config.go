package nmgr

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/tokamak-network/tokamak-trunks/utils"
)

type Config struct {
	DockerComposeDirPath string
	DeployConfigFilePath string
	AddressFilePath      string
	AllocFilePath        string
	FaucetAccounts       []common.Address
}

func NewConfig(cfg CLIConfig, accounts ...common.Address) *Config {
	return &Config{
		DockerComposeDirPath: utils.ConvertToAbsPath(cfg.DockerComposeDirPath),
		DeployConfigFilePath: utils.ConvertToAbsPath(cfg.DeployConfigFilePath),
		AddressFilePath:      utils.ConvertToAbsPath(cfg.AddressFilePath),
		AllocFilePath:        utils.ConvertToAbsPath(cfg.AllocFilePath),
		FaucetAccounts:       accounts,
	}
}
