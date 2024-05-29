package reporter

import (
	"github.com/tokamak-network/tokamak-trunks/utils"
	"github.com/urfave/cli/v2"
)

const (
	l2BlockTimeName = "l2-block-time"
)

type CLIConfig struct {
	L2BlockTime uint64
}

func CLIFlags(envPrefix string) []cli.Flag {
	return []cli.Flag{
		&cli.Uint64Flag{
			Name:    l2BlockTimeName,
			Usage:   "L2Block time",
			EnvVars: utils.PrefixEnvVars(envPrefix, "L2_BLOCK_TIME"),
		},
	}
}

func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		L2BlockTime: ctx.Uint64(l2BlockTimeName),
	}
}
