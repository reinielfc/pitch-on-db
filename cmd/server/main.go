package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"pitch-on-db/handler"
	"pitch-on-db/repository"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://pitchondb:pitchondb@localhost:5432/pitchondb?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	q := repository.New(db)
	h := handler.NewPigeonHandler(q)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /pigeons", h.List)
	mux.HandleFunc("GET /pigeons/{id}", h.Get)
	mux.HandleFunc("POST /pigeons", h.Create)

	addr := ":8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}
