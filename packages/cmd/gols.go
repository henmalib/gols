package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/urfave/cli/v3"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.IntN(len(letterRunes))]
	}
	return string(b)
}

func main() {
	validate := validator.New(validator.WithRequiredStructEnabled())

	cli := &cli.Command{
		Name:  "gols",
		Usage: "gols [url]",
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
			// TODO: deleting/editing already created links
		},
		Action: func(context context.Context, cmd *cli.Command) error {
			baseUrl := cmd.Args().First()
			if err := validate.Var(baseUrl, "uri,required"); err != nil {
				return err
			}

			shortUrl := cmd.String("short")

			if shortUrl == "" {
				shortUrl = RandStringRunes(rand.IntN(9) + 3)
			}

			body := []byte(fmt.Sprintf(`{
                    "short": "%s",
                "url": "%s"
                }`, shortUrl, baseUrl))

			r, err := http.NewRequest("POST", "http://localhost:5050/api/links", bytes.NewBuffer(body))
			if err != nil {
				return err
			}
			r.Header.Add("Content-Type", "application/json")
			r.Header.Add("Authorization", cmd.String("auth"))

			client := &http.Client{}
			res, err := client.Do(r)
			if err != nil {
				return err
			}

			defer res.Body.Close()

			bytes, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			fmt.Println(string(bytes))

			return nil
		},
	}

	if err := cli.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
