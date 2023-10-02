package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func getAddress(
	_ context.Context,
	serverAddr,
	domainName string,
) (string, error) {
	baseUrl, err := url.Parse(serverAddr)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Add("domainName", domainName)

	baseUrl.RawQuery = params.Encode()

	resp, err := http.Get(baseUrl.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return "", errors.Errorf(
			"Received err code %d with message: %s\n",
			resp.StatusCode,
			string(body),
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func listEntries(
	_ context.Context,
	serverAddr string,
) ([]string, error) {
	baseUrl, err := url.Parse(serverAddr)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(baseUrl.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.Errorf(
			"Received err code %d with message: %s\n",
			resp.StatusCode,
			string(body),
		)
	}

	scanner := bufio.NewScanner(resp.Body)

	ret := []string{}
	for scanner.Scan() {
		ret = append(ret, scanner.Text())
	}

	return ret, nil
}

func addEntry(
	ctx context.Context,
	serverAddr string,
	domainName,
	address string,
) error {
	baseUrl, err := url.Parse(serverAddr)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Add("domainName", domainName)
	params.Add("address", address)

	baseUrl.RawQuery = params.Encode()

	resp, err := http.Post(
		baseUrl.String(),
		"application/x-www-form-urlencoded",
		nil,
	)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.Errorf(
			"Received err code %d with message: %s\n",
			resp.StatusCode,
			string(body),
		)
	}

	return nil
}

func delEntry(
	ctx context.Context,
	serverAddr string,
	domainName string,
) error {
	baseUrl, err := url.Parse(serverAddr)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Add("domainName", domainName)

	baseUrl.RawQuery = params.Encode()

	req, err := http.NewRequest("DELETE", baseUrl.String(), nil)
	if err != nil {
		return err
	}

	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.Errorf(
			"Received err code %d with message: %s\n",
			resp.StatusCode,
			string(body),
		)
	}

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
