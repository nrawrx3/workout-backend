package backend

import (
	"fmt"
	"log"
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
	DB         *gorm.DB
	Cfg        *config.Config
	Router     *mux.Router
	HttpServer *http.Server
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

func (app *App) Init(cfg *config.Config) error {
	log.Printf("Init server...")

	// Set up stores
	userStore := store.NewUserStore(app.DB)

	// Router
	router := mux.NewRouter()

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

	app.HttpServer = &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("%s:%d", app.Cfg.Host, app.Cfg.TLSPort),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}
	app.Router = router
	return nil
}

func (app *App) RunServer(cfg *config.Config) error {
	if cfg.UseSelfSignedTLS {
		log.Printf("connect to https://localhost:%d/%s for GraphQL playground", app.Cfg.TLSPort, strings.TrimPrefix(constants.GqlPlaygroundApiPath, "/"))

		log.Printf("call https://localhost:%d/%s with GraphQL queries", app.Cfg.TLSPort, strings.TrimPrefix(constants.GqlQueryApiPath, "/"))

		return app.HttpServer.ListenAndServeTLS("./dev-certs/server.crt", "dev-certs/server.key")
	} else {
		log.Printf("connect to http://localhost:%d/%s for GraphQL playground", app.Cfg.Port, strings.TrimPrefix(constants.GqlPlaygroundApiPath, "/"))

		log.Printf("call http://localhost:%d/%s with GraphQL queries", app.Cfg.Port, strings.TrimPrefix(constants.GqlQueryApiPath, "/"))
		return app.HttpServer.ListenAndServe()
	}
}

func (app *App) RoutesSummary() {
	logger := log.New(os.Stdout, "", 0)
	err := app.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			logger.Println(pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			logger.Println("Path regexp:", pathRegexp)
		}
		// queriesTemplates, err := route.GetQueriesTemplates()
		// if err == nil {
		// 	logger.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		// }
		// queriesRegexps, err := route.GetQueriesRegexp()
		// if err == nil {
		// 	logger.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		// }
		methods, err := route.GetMethods()
		if err == nil {
			logger.Println("Methods:", strings.Join(methods, ","))
		}
		if v := reflect.ValueOf(route.GetHandler()); v.Kind() == reflect.Func {
			logger.Println("HandlerFn: ", runtime.FuncForPC(v.Pointer()).Name())
		}
		logger.Println()
		return nil
	})

	if err != nil {
		logger.Println(err)
	}
}
