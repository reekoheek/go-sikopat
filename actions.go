package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/reekoheek/go-sikopat/api"
	"gopkg.in/urfave/cli.v2"
)

type (
	action_t struct {
		api       *api.Api
		token     string
		tokenFile string
	}
)

func (a *action_t) profile(ctx *cli.Context) error {
	profile, err := a.api.Profile(a.token, false)
	if err != nil {
		return err
	}

	fmt.Printf("Username: %s\n", profile.Username)

	fmt.Println("\nSales:")
	for _, sale := range profile.Sales {
		fmt.Printf("%-20s %2s %3d %6d %6d %s\n", sale.Product, sale.Payment[:2], sale.Qty, sale.Price, sale.Total, sale.Date)
	}

	fmt.Printf("\nBalance:  %d\n", profile.Balance)
	return nil
}

func (a *action_t) buy(ctx *cli.Context) error {
	products, err := a.api.Products(a.token, ctx.Args().First())
	if err != nil {
		return err
	}

	found := len(products)
	if found > 1 {
		fmt.Printf("There is %d product found to buy\n\n", found)
		i := 0
		for _, product := range products {
			i++
			fmt.Printf("(%2d) %-20s %4d %7d\n", i, product.Name, product.Qty, product.Price)
		}
		fmt.Println("\nBuying aborted, to many product candidates")
		return nil
	}

	qty := ctx.Int("quantity")

	var product *api.Product
	for _, product = range products {
		break
	}
	fmt.Printf("Product:  %-20s %4d %7d\n", product.Name, product.Qty, product.Price)
	fmt.Printf("Quantity: %d\n", qty)

	if !ctx.Bool("force") {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Are you sure to buy this item (yes|no) ? ")
		text, _ := reader.ReadString('\n')
		res := strings.Trim(text, "\n\r\t ")
		if res != "yes" {
			fmt.Println("Cancel buying")
			return nil
		}
	}

	return a.api.Buy(a.token, product, qty)
}

func (a *action_t) search(ctx *cli.Context) error {
	products, err := a.api.Products(a.token, ctx.Args().First())
	if err != nil {
		return err
	}

	for _, product := range products {
		if product.Id != "" {
			fmt.Printf("%-30s %-20s %5d %20d\n", product.Id, product.Name, product.Qty, product.Price)
		}
	}

	return nil
}

func (a *action_t) login(ctx *cli.Context) error {
	var (
		reader   *bufio.Reader
		text     string
		username string
		password string
	)

	reader = bufio.NewReader(os.Stdin)
	fmt.Print("Username: ")
	text, _ = reader.ReadString('\n')
	username = strings.Trim(text, "\n\r\t ")

	reader = bufio.NewReader(os.Stdin)
	fmt.Print("Password: ")
	text, _ = reader.ReadString('\n')
	password = strings.Trim(text, "\n\r\t ")

	token, err := a.api.Login(&api.Profile{
		Username: username,
		Password: password,
	})
	if err != nil {
		return err
	}

	return a.persist(token)
}

func (a *action_t) logout(ctx *cli.Context) error {
	if err := a.api.Logout(a.token); err != nil {
		return err
	}

	return os.Remove(a.tokenFile)
}

func (a *action_t) persist(token string) error {
	if a.tokenFile == "" {
		return nil
	}

	contentBytes := []byte(token)
	return ioutil.WriteFile(a.tokenFile, contentBytes, 0644)
}

func newAction(ap *api.Api, tokenFile string) *action_t {
	token := ""

	if tokenFile != "" {
		content, err := ioutil.ReadFile(tokenFile)
		if err == nil {
			token = strings.Trim(string(content), "\r\n\t ")
		}
	}

	action := &action_t{
		api:       ap,
		tokenFile: tokenFile,
		token:     token,
	}

	return action
}
