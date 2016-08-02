package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/reekoheek/go-sikopat/api"

	"gopkg.in/urfave/cli.v2"
)

const (
	TOKEN_FILE = "sikopat.token"
	BASE_URL   = "https://sikopat.xinix.co.id/index.php"
)

func main() {
	log.SetFlags(log.Lshortfile)

	usr, _ := user.Current()

	configDir := os.Getenv("SIKOPAT_DIR")
	if os.Getenv("SIKOPAT_DIR") == "" && usr != nil {
		configDir = filepath.Join(usr.HomeDir, ".sikopat")
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			panic("Cannot create config dir at " + configDir)
		}
	}

	action := newAction(createApiImpl(configDir), configDir, TOKEN_FILE)

	app := &cli.App{
		Name:    "go-sikopat",
		Usage:   "Sikopat cli",
		Version: "0.2.0",
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
			{
				Name:   "info",
				Usage:  "application info",
				Action: action.info,
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

func createApiImpl(configDir string) *api.ApiImpl {
	a, err := api.New(BASE_URL, configDir, "")
	if err != nil {
		handleError(err)
	}

	return a
}
