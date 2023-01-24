package backend

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/nrawrx3/workout-backend/config"
	"github.com/nrawrx3/workout-backend/constants"
	"github.com/nrawrx3/workout-backend/graph"
	"github.com/nrawrx3/workout-backend/store"
	"gorm.io/gorm"

	"github.com/rs/cors"
)

type App struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func NewApp(cfg *config.Config) (*App, error) {
	gormDB, err := store.OpenGorm(cfg.Sqlite.SqliteDSN())
	if err != nil {
		return nil, err
	}

	return &App{
		DB:  gormDB,
		Cfg: cfg,
	}, nil
}

func (app *App) RunServer(cfg *config.Config) error {
	log.Printf("Running server...")

	allowedOrigins := append([]string{}, cfg.Cors.AllowedOrigins...)
	if cfg.Cors.AllowAll {
		allowedOrigins = append(allowedOrigins, "*")
	}

	log.Printf("cors allowed-origins: %+v", allowedOrigins)

	corsObject := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
	})

	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{
			DB: app.DB,
		},
	}))

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New(1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})

	http.Handle(constants.GqlQueryApiPath, corsObject.Handler(srv))
	http.Handle(constants.GqlPlaygroundApiPath, playground.Handler("GraphQL playground", constants.GqlQueryApiPath))

	log.Printf("connect to http://localhost:%d/%s for GraphQL playground", app.Cfg.Port, strings.TrimPrefix(constants.GqlPlaygroundApiPath, "/"))

	log.Printf("call http://localhost:%d/%s with GraphQL queries", app.Cfg.Port, strings.TrimPrefix(constants.GqlQueryApiPath, "/"))

	return http.ListenAndServe(fmt.Sprintf(":%d", app.Cfg.Port), nil)
}
