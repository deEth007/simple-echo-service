package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/tokamak-network/tokamak-trunks/account"
	"github.com/tokamak-network/tokamak-trunks/cmd/flags"
	"github.com/tokamak-network/tokamak-trunks/trunks"
)

func main() {
	app := cli.NewApp()
	app.Name = "tokamak-trunks"
	app.Usage = "tokamak overload test tool"
	app.Flags = flags.Flags
	app.Commands = []*cli.Command{
		{
			Name:   "start",
			Usage:  "start load test",
			Action: trunks.Main(),
		},
		{
			Name:  "account",
			Usage: "commands accounts for tx load test",
			Subcommands: []*cli.Command{
				{
					Name:   "generate",
					Usage:  "generate accounts",
					Flags:  account.GenerateCLIFlags("TOKAMAK_TRUNKS"),
					Action: account.Main(),
				},
				{
					Name:   "faucet",
					Usage:  "faucet ETH(TON at Thanos)",
					Flags:  account.FaucetCLIFlags("TOKAMAK_TRUNKS"),
					Action: account.Main(),
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func init() {
	homeDir, _ := os.UserHomeDir()
	trunksDir := fmt.Sprintf("%s/%s", homeDir, ".tokamak-trunks")
	if _, err := os.Stat(trunksDir); os.IsNotExist(err) {
		os.Mkdir(trunksDir, 0755)
	}
}
