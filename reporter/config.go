package reporter

import "math/big"

type Config struct {
	l2BlockTime *big.Int
	filename    string
}

func NewConfig(cfg CLIConfig) *Config {
	return &Config{
		l2BlockTime: new(big.Int).SetUint64(cfg.L2BlockTime),
		filename:    cfg.outputFileName,
	}
}
