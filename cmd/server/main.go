package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	db "github.com/rick-astral-cat/flizix-api/db/sqlc"
	"github.com/rick-astral-cat/flizix-api/internal/api"
	"github.com/rick-astral-cat/flizix-api/internal/config"
	_ "modernc.org/sqlite"
)

// @title	Flizix API
// @version	1.0
// @description	Backend for personal finances
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url	http://www.swagger.io/support
// @contact.email	support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	dbConn, err := sql.Open("sqlite", cfg.DbUrl)
	if err != nil {
		log.Fatalf("Can´t connect to flizix database: %v", err)
	}
	defer dbConn.Close()
	queries := db.New(dbConn)

	log.Println("### FLIZIX STARTING ON", cfg.AppEnv, " ###")
	log.Println("Database URL:", cfg.DbUrl)
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	apiCfg := &api.Config{
		Queries:   queries,
		JWTSecret: cfg.JWTSecret,
	}
	apiCfg.RegisterRoutes(mux, cfg.AppEnv)

	go func() {
		log.Printf("Listening on port %s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error on server %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down server gracefully...")
}
