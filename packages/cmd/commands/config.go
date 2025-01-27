package commands

import (
	"context"

	"github.com/henmalib/gols/packages/cmd/config"
	"github.com/henmalib/gols/packages/cmd/utils"
	"github.com/urfave/cli/v3"
)

func ChangeConfig() *cli.Command {
	return &cli.Command{
		Name:    "config",
		Aliases: []string{"cfg"},
		Usage:   "gols cfg --server http://localhost:5000 --key ajskldjaskl",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "server",
				Aliases:  []string{"s"},
				OnlyOnce: true,
				Usage:    "Provide a server url",
			},
			&cli.StringFlag{
				Name:     "apikey",
				Aliases:  []string{"key"},
				OnlyOnce: true,
				Usage:    "Auth key for the server",
			},
		},
		Action: changeConfigAction(),
	}
}

func changeConfigAction() func(context.Context, *cli.Command) error {
	return func(ctx context.Context, c *cli.Command) error {
		cfg, _ := config.ReadConfigFile()

		cfg.Server = utils.SelectFirstString(c.String("server"), cfg.Server)
		cfg.AuthKey = utils.SelectFirstString(c.String("apikey"), cfg.AuthKey)

		return config.WriteConfigFile(&cfg)
	}
}
