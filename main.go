package main

import (
	"fmt"
	"log"
	"os"

	"github.com/reekoheek/go-sikopat/api"

	"gopkg.in/urfave/cli.v2"
)

const (
	TOKEN_FILE = "sikopat.token"
	BASE_URL   = "http://sikopat.xinix.co.id/index.php"
)

func main() {
	log.SetFlags(log.Lshortfile)

	action := newAction(createApi(), TOKEN_FILE)

	app := &cli.App{
		Name:    "go-sikopat",
		Usage:   "Sikopat cli",
		Version: "0.1.0",
		Commands: []*cli.Command{
			{
				Name:   "profile",
				Usage:  "show user profile",
				Action: action.profile,
			},
			//{
			//	Name:   "sales",
			//	Usage:  "show sales",
			//	Action: action.sales,
			//},
			{
				Name:  "buy",
				Usage: "buy as user",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Value:   false,
						Usage:   "buy without confirmation",
					},
					&cli.IntFlag{
						Name:    "quantity",
						Aliases: []string{"n"},
						Value:   1,
						Usage:   "number of quantity",
					},
				},
				Action: action.buy,
			},
			{
				Name:   "logout",
				Usage:  "logout as user",
				Action: action.logout,
			},
			{
				Name:   "login",
				Usage:  "login as user",
				Action: action.login,
			},
			{
				Name:   "search",
				Usage:  "search items to buy",
				Action: action.search,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		handleError(err)
	}
}

func handleError(err error) {
	fmt.Fprintf(os.Stderr, "Error caught: %s\n", err.Error())
	os.Exit(1)
}

func createApi() *api.Api {
	a, err := api.New(BASE_URL, "")
	if err != nil {
		handleError(err)
	}

	return a
}
