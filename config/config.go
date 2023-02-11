package config

import (
	"bytes"
	"log"
	"strings"
	"text/template"

	config "github.com/golobby/config/v3"
	"github.com/golobby/config/v3/pkg/feeder"
)

type Config struct {
	Sqlite         SqliteConfig `json:"sqlite"`
	MigrationsPath string       `json:"migrations_path"`
	Host           string       `json:"host"`
	Port           int          `json:"port"`
	TLSPort        int          `json:"tls_port"`

	Cors struct {
		AllowedOrigins []string `json:"allowed_origins"`
		AllowAll       bool     `json:"allow_all"`
	} `json:"cors"`
	CookieSecretKey string `json:"cookie_secret_key"`
	CookieName      string `json:"cookie_name"`
	CookieDomain    string `json:"cookie_domain"`

	// For testing purposes. In production, use a SSL reverse proxy instead.
	UseSelfSignedTLS bool `json:"use_self_signed_tls"`
}

type SqliteConfig struct {
	File         string `json:"file"`
	InMemoryMode bool   `json:"in_memory"`
	ReadOnlyMode bool   `json:"read_only"`
}

const sqliteURLTemplate = `sqlite3://{{.DbFile}}`
const sqliteDSNTemplate = `file:{{.DbFile}}?{{.OptionsQueryParams}}`

// DB URL is required by the golang-migrate
func (c *SqliteConfig) SqliteDatabaseURL() string {
	tmpl, err := template.New("sqlite_db_url").Parse(sqliteURLTemplate)
	if err != nil {
		log.Fatalf("failed to parse sqliteURLTemplate: %v", err)
	}

	var tmplValues struct {
		DbFile string
	}
	tmplValues.DbFile = c.File
	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, &tmplValues)
	if err != nil {
		log.Fatal(err)
	}
	return buf.String()
}

// DSN is required by the sqlite3 driver to open connection
func (c *SqliteConfig) SqliteDSN() string {
	tmpl, err := template.New("sqlite_dsn").Parse(sqliteDSNTemplate)
	if err != nil {
		log.Fatalf("failed to parse sqliteDSNTemplate: %v", err)
	}

	queryParams := make(map[string][]string)
	queryParams["_foreign_keys"] = []string{"true"} // Enable foreign keys support

	queryParams["mode"] = []string{} // Set mode as per config

	if c.InMemoryMode {
		queryParams["mode"] = append(queryParams["mode"], "memory")
	}

	if c.ReadOnlyMode {
		queryParams["mode"] = append(queryParams["mode"], "ro")
	}

	var queryParamsString strings.Builder
	for k, values := range queryParams {
		for i, v := range values {
			queryParamsString.WriteString(k)
			queryParamsString.WriteRune('=')
			queryParamsString.WriteString(v)
			if i < len(values)-1 {
				queryParamsString.WriteRune('&')
			}
		}
	}

	var tmplValues struct {
		DbFile             string
		OptionsQueryParams string
	}

	tmplValues.DbFile = c.File
	tmplValues.OptionsQueryParams = queryParamsString.String()

	buf := bytes.NewBufferString("")
	tmpl.Execute(buf, &tmplValues)
	if err != nil {
		log.Fatal(err)
	}

	dsnString := buf.String()
	dsnString = strings.TrimSuffix(dsnString, "?")
	return dsnString
}

func (cfg *Config) LoadFromJSONFile(configFile string) error {
	jsonFeeder := feeder.Json{Path: configFile}
	c := config.New().AddFeeder(jsonFeeder).AddStruct(cfg)
	err := c.Feed()
	if err != nil {
		return err
	}
	return nil
}
