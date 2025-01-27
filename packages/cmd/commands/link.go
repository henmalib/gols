package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/henmalib/gols/packages/cmd/config"
	"github.com/henmalib/gols/packages/cmd/utils"

	"github.com/urfave/cli/v3"
)

func CreateLink() *cli.Command {
	return &cli.Command{
		Name:    "link",
		Aliases: []string{"l"},
		Usage:   "gols link [url]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "short",
				Aliases:  []string{"s", "sh"},
				OnlyOnce: true,
				Usage:    "Provide a custom short link",
			},
			&cli.StringFlag{
				Name:     "auth",
				OnlyOnce: true,
				Usage:    "Provide AUTH token of the server",
				Sources:  cli.EnvVars("AUTH_TOKEN"),
			},
			&cli.StringFlag{
				Name:     "server",
				OnlyOnce: true,
				Usage:    "Change server url",
				Sources:  cli.EnvVars("SERVER"),
			},
		},
		Action: createAction(),
	}
}

func createAction() func(context.Context, *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		cfg, err := config.ReadConfigFile()
		if err != nil {
			return errors.New("Config file is not created yet")
		}
		baseUrl := cmd.Args().First()
		if err := utils.Validate.Var(baseUrl, "uri,required"); err != nil {
			return err
		}

		shortUrl := cmd.String("short")
		body := []byte(fmt.Sprintf(`{
                    "short": "%s",
                    "url": "%s"
                }`, shortUrl, baseUrl))

		server := utils.SelectFirstString(cmd.String("server"), cfg.Server)
		r, err := http.NewRequest("POST", fmt.Sprintf("%s/api/links", server), bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		r.Header.Add("Content-Type", "application/json")
		r.Header.Add("Authorization", utils.SelectFirstString(cmd.String("auth"), cfg.AuthKey))

		client := &http.Client{}
		res, err := client.Do(r)
		if err != nil {
			return err
		}

		defer res.Body.Close()

		serverResponse, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		fmt.Println(string(serverResponse))

		return nil
	}

}
