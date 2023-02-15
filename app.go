package backend

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"github.com/nrawrx3/uno-backend/config"
	"github.com/nrawrx3/uno-backend/constants"
	"github.com/nrawrx3/uno-backend/graph"
	bk_handler "github.com/nrawrx3/uno-backend/handler"
	"github.com/nrawrx3/uno-backend/handler/middleware"
	"github.com/nrawrx3/uno-backend/model"
	"github.com/nrawrx3/uno-backend/store"
	"github.com/nrawrx3/uno-backend/util"
	"gorm.io/gorm"

	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type App struct {
	DB         *gorm.DB
	Cfg        *config.Config
	Router     *mux.Router
	HttpServer *http.Server
}

func NewApp(cfg *config.Config) (*App, error) {
	initGlobalLogger(cfg)
	gormDB, err := store.OpenGorm(cfg.Sqlite.SqliteDSN())
	if err != nil {
		return nil, err
	}

	return &App{
		DB:  gormDB,
		Cfg: cfg,
	}, nil
}

func initGlobalLogger(cfg *config.Config) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	if cfg.Logger.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
	zerolog.TimestampFieldName = cfg.Logger.TimestampFormat
	zerolog.TimeFieldFormat = util.ISO8601LayoutWithoutT
}

func (app *App) Init(cfg *config.Config) error {
	log.Info().Msg("Init server")

	// Set up stores
	userStore := store.NewUserStore(app.DB)

	// Router
	router := mux.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recover)

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

	log.Debug().Dict("creating cors-middleware", zerolog.Dict().Str("cors-allowed-origins", strings.Join(allowedOrigins, ",")))

	corsObject := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"Origin", "Accept", "Content-Type", "X-Requested-With"},
		// AllowOriginFunc: func(origin string) bool {
		// 	log.Printf("received origin: %s", origin)
		// 	return origin == "http://localhost:5180"
		// },
	})

	cookieInfo := model.SessionCookieInfo{
		CookieName: cfg.CookieName,
		Secure:     true,
		SecretKey:  cfg.CookieSecretKey,
		SameSite:   http.SameSiteNoneMode,
		Expires:    time.Now().Add(1 * time.Hour),
		HttpOnly:   true,
		Domain:     cfg.CookieDomain,
	}

	sessionCheckMiddle := middleware.NewSessionChecker(userStore, cookieInfo, aesCipher)

	loginHandler := bk_handler.NewLoginHandler(userStore, cookieInfo, aesCipher)

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

	gqlSubRouter := router.PathPrefix(constants.GqlRootApiPrefix).Subrouter()

	gqlSubRouter.Path(constants.GqlPlaygroundApiPath).HandlerFunc(playground.Handler("GraphQL playground", constants.GqlQueryApiPath))

	router.Path(constants.LoginPath).Handler(corsObject.Handler(http.HandlerFunc(loginHandler.Login)))

	gqlSubRouter.Path(constants.GqlQueryApiPath).Handler(
		corsObject.Handler(sessionCheckMiddle.Handler(srv)))

	router.Path(constants.AmILoggedInPath).Handler(
		corsObject.Handler(http.HandlerFunc(loginHandler.AmILoggedIn)))

	workoutsListHandler := bk_handler.NewWorkoutsListHandler(userStore)

	router.Path(constants.WorkoutsListPath).Methods("GET").Handler(corsObject.Handler(sessionCheckMiddle.Handler(
		http.HandlerFunc(workoutsListHandler.HandleGetWorkoutsList))))

	host := app.Cfg.Host
	if host == "" {
		host = "localhost"
	}

	app.Router = router
	return nil
}

func (app *App) RunServer(cfg *config.Config) error {
	if cfg.UseSelfSignedTLS {
		listenAddr := fmt.Sprintf("%s:%d", app.Cfg.Host, app.Cfg.TLSPort)
		log.Info().Str("listenAddr", listenAddr).Msg("starting http server")

		app.HttpServer = &http.Server{
			Handler:      app.Router,
			Addr:         listenAddr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  1 * time.Minute,
		}
		log.Info().Dict("gql-urls", zerolog.Dict().
			Str("Playground", fmt.Sprintf("https://localhost:%d/gql/playground", app.Cfg.TLSPort)).
			Str("Query", fmt.Sprintf("https://localhost:%d/gql/query", app.Cfg.TLSPort))).Msg("")

		return app.HttpServer.ListenAndServeTLS("./dev-certs/server.crt", "dev-certs/server.key")
	} else {
		listenAddr := fmt.Sprintf("%s:%d", app.Cfg.Host, app.Cfg.Port)
		log.Info().Str("listenAddr", listenAddr).Msg("starting http server")

		app.HttpServer = &http.Server{
			Handler:      app.Router,
			Addr:         listenAddr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  1 * time.Minute,
		}

		log.Info().Dict("gql-urls", zerolog.Dict().
			Str("Playground", fmt.Sprintf("https://localhost:%d/gql/playground", app.Cfg.Port)).
			Str("Query", fmt.Sprintf("https://localhost:%d/gql/query", app.Cfg.Port)))
		return app.HttpServer.ListenAndServe()
	}
}

func (app *App) RoutesSummary() {
	var sb strings.Builder

	err := app.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			sb.WriteString(pathTemplate)
			sb.WriteString("\n")
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			sb.WriteString("Path regexp:")
			sb.WriteString(pathRegexp)
			sb.WriteString("\n")
		}
		methods, err := route.GetMethods()
		if err == nil {
			sb.WriteString("Methods:")
			sb.WriteString(strings.Join(methods, ","))
			sb.WriteString("\n")
		}
		if v := reflect.ValueOf(route.GetHandler()); v.Kind() == reflect.Func {
			sb.WriteString("FirstHandler:")
			sb.WriteString(runtime.FuncForPC(v.Pointer()).Name())
			sb.WriteString("\n")
		}
		sb.WriteString("---\n")
		return nil
	})

	if err != nil {
		log.Panic().Err(err)
	} else {
		fmt.Print(sb.String())
	}
}
