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
	_ "modernc.org/sqlite"
)

func main() {
	dbConn, err := sql.Open("sqlite", "./flizix.db")
	if err != nil {
		log.Fatal("Can´t connect to flizix database: ", err)
	}
	defer dbConn.Close()
	queries := db.New(dbConn)

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	apiCfg := &api.Config{Queries: queries}
	apiCfg.RegisterRoutes(mux)

	go func() {
		log.Printf("Listening on port %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error on server %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down server gracefully...")
}
