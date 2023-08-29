// legitima runs the service.
// For details on how to configure it just run:
//
//	legitima --help
package main

import (
	"net/http"
	"os"
	"time"

	"github.com/birdie-ai/golibs/slog"
	"github.com/birdie-ai/legitima/api"
	"github.com/birdie-ai/legitima/mysql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Config holds the configuration for the service.
type Config struct {
	ClientID     string
	ClientSecret string
	HTTPAddr     string
}

func main() {
	logcfg, err := slog.LoadConfig("LEGITIMA")
	if err != nil {
		slog.Fatal("failed to load config", "error", err.Error())
	}

	err = slog.Configure(logcfg)
	if err != nil {
		slog.Fatal("failed to configure logger", "error", err.Error())
	}

	cfg := &Config{
		ClientID:     os.Getenv("LEGITIMA_GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("LEGITIMA_GOOGLE_CLIENT_SECRET"),
		HTTPAddr:     getEnvWithDefault("LEGITIMA_HTTP_ADDR", ":8080"),
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		slog.Fatal("missing google auth client id or secret")
	}

	googleOAuthConfig := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8080/callback",
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile"},
	}

	dbConfig := mysql.Config{
		URL:             os.Getenv("LEGITIMA_MYSQL_URL"),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	db, err := mysql.OpenDB(dbConfig)
	if err != nil {
		slog.Fatal("failed to open db", "error", err.Error())
	}

	storage := mysql.NewStorage(db)

	mux := http.NewServeMux()
	api.SetupAuth(mux, &googleOAuthConfig, storage)
	mux.HandleFunc("/", api.Home)
	api.SetupProfile(mux, storage)

	svr := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	slog.Info("starting server", "addr", svr.Addr)
	err = svr.ListenAndServe()
	if err != nil {
		slog.Fatal("server error", "error", err.Error())
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
