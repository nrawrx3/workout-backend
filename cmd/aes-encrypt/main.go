package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/nrawrx3/uno-backend/util"
	"github.com/urfave/cli/v2"
)

var cliFlags struct {
	hexKey       string
	dataString   string
	encodeBase64 bool
}

func main() {
	app := cli.App{
		Name:  "encrypt-with-aes-gcm",
		Usage: "encrypt string with AES-GCM",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "key",
				Aliases:     []string{"k"},
				Usage:       "hex key (64 chars, 256 bit)",
				Required:    true,
				Destination: &cliFlags.hexKey,
			},
		},

		Commands: []*cli.Command{
			{
				Name:    "encrypt",
				Usage:   "encrypt bytes",
				Aliases: []string{"enc", "e"},

				Action: func(ctx *cli.Context) error {
					aesCipher, err := util.NewAESCipher(cliFlags.hexKey)
					if err != nil {
						return err
					}
					encBytes, err := aesCipher.Encrypt([]byte(cliFlags.dataString))
					if err != nil {
						return err
					}

					if cliFlags.encodeBase64 {
						fmt.Printf("%s\n", base64.URLEncoding.EncodeToString(encBytes))
					} else {
						fmt.Printf("%s\n", hex.EncodeToString(encBytes))
					}
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "string",
						Aliases:     []string{"s"},
						Usage:       "string to encrypt",
						Required:    true,
						Destination: &cliFlags.dataString,
					},
					&cli.BoolFlag{
						Name:        "base64",
						Usage:       "encode encrypted bytes as base64, by default encoded as hex",
						Destination: &cliFlags.encodeBase64,
					},
				},
			},
			{
				Name:    "decrypt",
				Usage:   "decrypt bytes",
				Aliases: []string{"dec", "d"},

				Action: func(ctx *cli.Context) error {
					aesCipher, err := util.NewAESCipher(cliFlags.hexKey)
					if err != nil {
						return err
					}

					var encryptedBytes []byte

					if cliFlags.encodeBase64 {
						bytes, err := base64.URLEncoding.DecodeString(cliFlags.dataString)
						if err != nil {
							return nil
						}
						encryptedBytes = bytes
					} else {
						bytes, err := hex.DecodeString(cliFlags.dataString)
						if err != nil {
							return nil
						}
						encryptedBytes = bytes
					}

					decryptedBytes, err := aesCipher.Decrypt(encryptedBytes)
					if err != nil {
						return err
					}
					fmt.Printf("%s", string(decryptedBytes))
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "string",
						Aliases:     []string{"s"},
						Usage:       "string to decrypt",
						Required:    true,
						Destination: &cliFlags.dataString,
					},
					&cli.BoolFlag{
						Name:        "base64",
						Usage:       "indicate that the encrypted string is in base64, as opposed to hex by default",
						Destination: &cliFlags.encodeBase64,
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
