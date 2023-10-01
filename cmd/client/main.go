package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func getAddress(
	ctx context.Context,
	serverAddr,
	domainName string,
) (string, error) {
	return "", nil
}

func listEntries(
	ctx context.Context,
	serverAddr string,
) ([]string, error) {
	return nil, nil
}

func addEntry(
	ctx context.Context,
	serverAddr string,
	domainName,
	address string,
) error {
	return nil
}

func delEntry(
	ctx context.Context,
	serverAddr string,
	domainName string,
) error {
	return nil
}

func main() {
	app := &cli.App{
		Name:  "tolink-admin",
		Usage: "Admin cli for managing the tolink server",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "server",
				Aliases: []string{"s"},
				Value:   "127.0.0.1:8091",
				Usage:   "tolink admin server address",
			},
		},

		Commands: []*cli.Command{

			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "Get the address of the give domain name.",
				Action: func(c *cli.Context) error {
					dn := c.Args().First()
					svrAdr := c.String("server")
					addr, err := getAddress(context.TODO(), svrAdr, dn)
					if err != nil {
						return err
					}
					fmt.Println(addr)
					return nil
				},
			},

			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List all domainName:address entries.",
				Action: func(c *cli.Context) error {
					svrAdr := c.String("server")
					entries, err := listEntries(context.TODO(), svrAdr)
					if err != nil {
						return err
					}
					for _, entry := range entries {
						fmt.Println(entry)
					}
					return nil
				},
			},

			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "say hello",
				Action: func(c *cli.Context) error {
					dn := c.Args().First()
					adr := c.Args().Get(1)
					svrAdr := c.String("server")
					if dn == "" || adr == "" || svrAdr == "" {
						return errors.New("required argument(s) not provided")
					}
					return addEntry(context.TODO(), svrAdr, dn, adr)
				},
			},

			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "say hello",
				Action: func(c *cli.Context) error {
					dn := c.Args().First()
					svrAdr := c.String("server")
					if dn == "" || svrAdr == "" {
						return errors.New("required argument(s) not provided")
					}
					return delEntry(context.TODO(), svrAdr, dn)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}

}
