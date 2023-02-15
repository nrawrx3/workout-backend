package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	cli "github.com/urfave/cli/v2"
	"gorm.io/gorm"

	backend "github.com/nrawrx3/uno-backend"
	"github.com/nrawrx3/uno-backend/config"
	"github.com/nrawrx3/uno-backend/graph"
	"github.com/nrawrx3/uno-backend/store"
)

const defaultGQLPlaygroundPort = "8080"

var cliFlags struct {
	configFile string
}

func main() {
	log.Default().SetFlags(log.Lshortfile | log.Ltime)

	configFlag := cli.StringFlag{
		Name:        "config",
		Aliases:     []string{"c"},
		Usage:       "Load configuration from `FILE`",
		Destination: &cliFlags.configFile,
		Required:    true,
	}

	app := &cli.App{
		Name:  "workout-backend",
		Usage: "a simple graphql backend for workout app",

		Commands: []*cli.Command{
			{
				Name:  "migrate",
				Usage: "run migrations",
				Action: func(c *cli.Context) error {
					log.Default().SetFlags(0)
					log.Printf("Running migrations...")
					var cfg config.Config
					err := cfg.LoadFromJSONFile(cliFlags.configFile)
					if err != nil {
						return err
					}

					return store.RunDatabaseMigrations(&cfg)
				},
				Flags: []cli.Flag{&configFlag},
			},
			{
				Name:  "rollback",
				Usage: "rollback by one step",
				Action: func(c *cli.Context) error {
					var cfg config.Config
					err := cfg.LoadFromJSONFile(cliFlags.configFile)
					if err != nil {
						return err
					}
					return store.RunDatabaseRollback(&cfg)
				},
				Flags: []cli.Flag{&configFlag},
			},
			{
				Name:  "seed",
				Usage: "seed the database with test data",
				Flags: []cli.Flag{&configFlag},
				Action: func(c *cli.Context) error {
					var cfg config.Config
					err := cfg.LoadFromJSONFile(cliFlags.configFile)
					if err != nil {
						return err
					}

					db, err := store.OpenGorm(cfg.Sqlite.SqliteDSN())
					if err != nil {
						return err
					}

					tx := db.Begin()
					defer tx.Rollback()
					err = backend.SeedDatabase(db)
					if err == nil {
						tx.Commit()
						return nil
					}
					return err
				},
			},
			{
				Name:  "server",
				Usage: "run server",
				Flags: []cli.Flag{&configFlag},
				Action: func(c *cli.Context) error {
					cfg := new(config.Config)
					err := cfg.LoadFromJSONFile(cliFlags.configFile)
					if err != nil {
						return err
					}

					app, err := backend.NewApp(cfg)
					if err != nil {
						return err
					}
					app.Init(cfg)
					return app.RunServer(cfg)
				},
			},
			{
				Name:  "routes",
				Usage: "print routes",
				Flags: []cli.Flag{&configFlag},
				Action: func(c *cli.Context) error {
					cfg := new(config.Config)
					err := cfg.LoadFromJSONFile(cliFlags.configFile)
					if err != nil {
						return err
					}

					app, err := backend.NewApp(cfg)
					if err != nil {
						return err
					}
					app.Init(cfg)
					app.RoutesSummary()
					return nil
				},
			},
			{
				Name:  "gql-playground",
				Usage: "run gql playground server",
				Flags: []cli.Flag{&configFlag},
				Action: func(c *cli.Context) error {
					var cfg config.Config
					err := cfg.LoadFromJSONFile(cliFlags.configFile)
					if err != nil {
						log.Fatal(err)
					}

					db, err := store.OpenGorm(cfg.Sqlite.SqliteDSN())
					if err != nil {
						return err
					}

					return startGQLPlayground(db, &cfg)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func startGQLPlayground(db *gorm.DB, cfg *config.Config) error {
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%d/ for GraphQL playground", cfg.Port)
	return http.ListenAndServe(fmt.Sprintf("localhost:%d", cfg.Port), nil)
}
