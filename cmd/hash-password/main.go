package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nrawrx3/workout-backend/util"
	"github.com/urfave/cli/v2"
)

var cliFlags struct {
	password           string
	base64passwordHash string
}

func main() {
	app := cli.App{
		Name:  "password-hash",
		Usage: "generate and compare base64 password hashes with bcrypt",

		Commands: []*cli.Command{
			{
				Name:  "generate",
				Usage: "generate password hash with bcrypt",
				Action: func(ctx *cli.Context) error {
					hash, err := util.HashPasswordBase64(cliFlags.password)
					if err != nil {
						log.Printf("failed to generate hash: %v", err)
						return err
					}
					fmt.Printf("%s\n", hash)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "password",
						Aliases:     []string{"p"},
						Usage:       "password",
						Required:    true,
						Destination: &cliFlags.password,
					},
				},
			},
			{
				Name:  "compare",
				Usage: "compare password with a hash",
				Action: func(ctx *cli.Context) error {
					match, err := util.PasswordMatchesHash(cliFlags.password, cliFlags.base64passwordHash)
					if err != nil {
						log.Print(err)
						return err
					}
					if match {
						fmt.Print("Match")
					} else {
						fmt.Print("No match")
					}
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "password",
						Aliases:     []string{"p"},
						Usage:       "password",
						Required:    true,
						Destination: &cliFlags.password,
					},
					&cli.StringFlag{
						Name:        "hash-bytes",
						Aliases:     []string{"b"},
						Usage:       "base64 hash to compare against",
						Required:    true,
						Destination: &cliFlags.base64passwordHash,
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
