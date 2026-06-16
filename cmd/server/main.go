package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"forum/internal/store"
	"forum/internal/web"
)

func main() {
	dbPath := env("FORUM_DB", "data/forum.db")
	addr := env("FORUM_ADDR", ":8080")

	st, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("base de données : %v", err)
	}
	defer st.Close()

	app, err := web.New(st)
	if err != nil {
		log.Fatalf("application : %v", err)
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      app.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Forum des 4 Couleurs ▸ http://localhost%s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
