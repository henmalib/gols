package main

import (
	"context"
	"log"
	"os"

	"github.com/henmalib/gols/packages/cmd/commands"
	"github.com/urfave/cli/v3"
)

func main() {
	cli := &cli.Command{
		Name: "gols",
		Commands: []*cli.Command{
			commands.CreateLink(),
			commands.ChangeConfig(),
		},
	}

	if err := cli.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
