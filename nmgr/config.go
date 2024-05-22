package nmgr

import (
	"github.com/tokamak-network/tokamak-trunks/utils"
)

type Config struct {
	DockerComposeFilePath string
	L1GenesisFilePath     string
	L2GenesisFilePath     string
	RollupConfigFilePath  string
	AddressFilePath       string
	JwtFilePath           string
}

func NewConfig(cfg CLIConfig) *Config {
	return &Config{
		DockerComposeFilePath: utils.ConvertToAbsPath(cfg.DockerComposeFilePath),
		L1GenesisFilePath:     utils.ConvertToAbsPath(cfg.L1GenesisFilePath),
		L2GenesisFilePath:     utils.ConvertToAbsPath(cfg.L2GenesisFilePath),
		RollupConfigFilePath:  utils.ConvertToAbsPath(cfg.RollupConfigFilePath),
		AddressFilePath:       utils.ConvertToAbsPath(cfg.AddressFilePath),
		JwtFilePath:           utils.ConvertToAbsPath(cfg.JwtSecretFilePath),
	}
}
