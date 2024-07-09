package account

import (
	"github.com/urfave/cli/v2"

	"github.com/tokamak-network/tokamak-trunks/utils"
)

const (
	DistributorPrivateKeyName = "distributor-private-key"
	RpcURLName                = "rpc-url"
	CountAccountsName         = "count-accounts"
)

type CLIConfig struct {
	DistributorPrivateKey string
	RpcURL                string
	CountAccounts         int64
}

func GenerateCLIFlags(envPrefix string) []cli.Flag {
	return []cli.Flag{
		&cli.Int64Flag{
			Name:    CountAccountsName,
			Usage:   "number of accounts to generate",
			EnvVars: utils.PrefixEnvVars(envPrefix, "COUNT_ACCOUNTS"),
		},
	}
}

func FaucetCLIFlags(envPrefix string) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    DistributorPrivateKeyName,
			Usage:   "private key for ETH(TON) distribute",
			EnvVars: utils.PrefixEnvVars(envPrefix, "DISTRIBUTOR_PRIVATE_KEY"),
		},
		&cli.StringFlag{
			Name:    RpcURLName,
			Usage:   "RPC URL",
			EnvVars: utils.PrefixEnvVars(envPrefix, "RPC_URL"),
		},
	}
}

func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		DistributorPrivateKey: ctx.String(DistributorPrivateKeyName),
		CountAccounts:         ctx.Int64(CountAccountsName),
		RpcURL:                ctx.String(RpcURLName),
	}
}
