package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/henmalib/gols/packages/web/env"
	"github.com/henmalib/gols/packages/web/handlers"
	"github.com/henmalib/gols/packages/web/helpers"
	_ "github.com/mattn/go-sqlite3"
)

type Url struct {
	Id    int
	Url   string
	Short string

	CreatedAt time.Time
}

type CreateLinkPayload struct {
	Url   string `json:"url" validate:"required,uri"`
	Short string `json:"short"`
}

var validate = validator.New(validator.WithRequiredStructEnabled())

func EnsureDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:./database/db.sqlite?cache=shared&mode=rwc")

	if err != nil {
		return db, err
	}
	db.SetMaxOpenConns(1)

	if err = db.Ping(); err != nil {
		return db, err
	}

	// TODO: usage count
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            short TEXT UNIQUE NOT NULL,
            original TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );`)

	if err != nil {
		return db, err
	}

	return db, nil
}

func RandStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.IntN(len(letterRunes))]
	}
	return string(b)
}

func main() {
	mux := http.NewServeMux()

	db, err := EnsureDB()
	if err != nil {
		log.Fatalln(err)
	}

	app := handlers.App{
		DB: db,
	}

	mux.HandleFunc("POST /api/links", func(w http.ResponseWriter, r *http.Request) {
		authKey := r.Header.Get("Authorization")
		if authKey != env.Env.ApiKey {
			helpers.WriteError(w, http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError)
		}

		payload := CreateLinkPayload{}

		if err = json.Unmarshal(body, &payload); err != nil {
			helpers.WriteError(w, http.StatusBadRequest)
			log.Println(err)
			return
		}

		if err = validate.Struct(payload); err != nil {
			helpers.WriteError(w, http.StatusBadRequest)
			log.Println(err)
			return
		}

		if payload.Short == "" {
			payload.Short = RandStringRunes(rand.IntN(9) + 3)
		}

		_, err = db.Exec(`INSERT INTO urls (short, original) VALUES (?, ?);`, "/"+payload.Short, payload.Url)
		if err != nil {
			http.Error(w, "Short link with this url already exists", http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, "http://localhost:5050/%s", payload.Short)
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		row := db.QueryRow("SELECT original FROM urls WHERE short = ?", r.URL.String())
		url := &Url{}

		err := row.Scan(&url.Url)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helpers.WriteError(w, http.StatusNotFound)
				return
			}

			log.Println(err)
			helpers.WriteError(w, http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, url.Url, http.StatusMovedPermanently)
	})

	mux.HandleFunc("DELETE /api/links", app.DeleteLinkHandler)

	server := http.Server{
		Addr:    env.Env.Host,
		Handler: mux,
	}

	log.Printf("Server is starting: %s", env.Env.Host)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
