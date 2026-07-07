package main

import (
	"context"
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
// @BasePath  /api
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

	//Seed default system data
	ctx := context.Background()
	if err = api.SeedDefaultAccountTypes(ctx, queries); err != nil {
		log.Fatalf("Error seeding default account types: %v", err)
	}

	userH := api.NewUserHandler(queries)
	authH := api.NewAuthHandler(queries, cfg.JWTSecret, cfg.TelegramBotToken, cfg.AppTLS)
	midH := api.NewMiddlewareHandler(authH, cfg.EnableCORS, cfg.AllowedOrigins)
	cardH := api.NewCardHandler(queries)
	accH := api.NewAccountHandler(queries)

	log.Println("### FLIZIX STARTING ON", cfg.AppEnv, " ###")
	log.Println("Database URL:", cfg.DbUrl)

	mux := http.NewServeMux()
	api.RegisterRoutes(mux, cfg.AppEnv, userH, authH, midH, cardH, accH)
	mainMux := http.NewServeMux()
	mainMux.Handle("/api/", http.StripPrefix("/api", mux))
	handleWithCORS := midH.CORSMiddleware(mainMux)
	
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      handleWithCORS,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

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
