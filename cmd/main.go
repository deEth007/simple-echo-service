package main

import (
	"log"
	"os"

	"github.com/tokamak-network/tokamak-trunks/cmd/flags"
	"github.com/tokamak-network/tokamak-trunks/trunks"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "tokamak-trunks"
	app.Usage = "tokamak overload test tool"
	app.Flags = flags.Flags
	app.Action = trunks.Main()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
