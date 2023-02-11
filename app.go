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
	bk_handler "github.com/nrawrx3/workout-backend/handler"
	"github.com/nrawrx3/workout-backend/handler/middleware"
	"github.com/nrawrx3/workout-backend/model"
	"github.com/nrawrx3/workout-backend/store"
	"github.com/nrawrx3/workout-backend/util"
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

	// Set up stores
	userStore := store.NewUserStore(app.DB)

	// Cipher we use for cookies
	aesCipher, err := util.NewAESCipher(cfg.CookieSecretKey)
	if err != nil {
		return err
	}

	// Set up CORS middleware
	allowedOrigins := append([]string{}, cfg.Cors.AllowedOrigins...)
	if cfg.Cors.AllowAll {
		allowedOrigins = append(allowedOrigins, "*")
	}

	log.Printf("cors allowed-origins: %+v", allowedOrigins)

	corsObject := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"Origin", "Accept", "Content-Type", "X-Requested-With"},
		// AllowOriginFunc: func(origin string) bool {
		// 	log.Printf("received origin: %s", origin)
		// 	return origin == "http://localhost:5180"
		// },
	})

	// Set up GraphQL handler
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

	http.Handle(constants.GqlPlaygroundApiPath, playground.Handler("GraphQL playground", constants.GqlQueryApiPath))

	cookieInfo := model.SessionCookieInfo{
		CookieName: cfg.CookieName,
		Secure:     true,
		SecretKey:  cfg.CookieSecretKey,
		SameSite:   http.SameSiteNoneMode,
		Expires:    time.Now().Add(1 * time.Hour),
		HttpOnly:   true,
		Domain:     cfg.CookieDomain,
	}

	sessionRedirHandler := middleware.NewSessionRedirectToLogin(userStore, cookieInfo, aesCipher)

	loginHandler := bk_handler.NewLoginHandler(userStore, cookieInfo, aesCipher)
	http.Handle("/login", corsObject.Handler(http.HandlerFunc(loginHandler.Login)))

	http.Handle(constants.GqlQueryApiPath,
		corsObject.Handler(sessionRedirHandler.Handler(srv)))

	http.Handle(constants.AmILoggedInPath,
		corsObject.Handler(http.HandlerFunc(loginHandler.AmILoggedIn)))

	workoutsListHandler := bk_handler.NewWorkoutsListHandler(userStore)
	http.Handle("/workouts",
		corsObject.Handler(
			sessionRedirHandler.Handler(
				http.HandlerFunc(workoutsListHandler.HandleGetWorkoutsList))))

	if cfg.UseSelfSignedTLS {
		log.Printf("connect to https://localhost:%d/%s for GraphQL playground", app.Cfg.TLSPort, strings.TrimPrefix(constants.GqlPlaygroundApiPath, "/"))

		log.Printf("call https://localhost:%d/%s with GraphQL queries", app.Cfg.TLSPort, strings.TrimPrefix(constants.GqlQueryApiPath, "/"))

		return http.ListenAndServeTLS(fmt.Sprintf(":%d", app.Cfg.TLSPort), "./dev-certs/server.crt", "dev-certs/server.key", nil)
	} else {
		log.Printf("connect to http://localhost:%d/%s for GraphQL playground", app.Cfg.Port, strings.TrimPrefix(constants.GqlPlaygroundApiPath, "/"))

		log.Printf("call http://localhost:%d/%s with GraphQL queries", app.Cfg.Port, strings.TrimPrefix(constants.GqlQueryApiPath, "/"))
		return http.ListenAndServe(fmt.Sprintf(":%d", app.Cfg.Port), nil)
	}
}
