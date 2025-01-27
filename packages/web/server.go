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
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/mattn/go-sqlite3"
)

type Url struct {
	Id    int
	Url   string
	Short string

	CreatedAt time.Time
}

func WriteError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
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

func getEnv() (string, string, error) {
	host := os.Getenv("HOST")
	apiKey := os.Getenv("API_KEY")

	if host == "" {
		host = ":5050"
	}
	if apiKey == "" {
		return host, apiKey, errors.New("No API_KEY provided")
	}

	return host, apiKey, nil
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

	host, apiKey, err := getEnv()
	if err != nil {
		log.Fatalln(err)
	}

	db, err := EnsureDB()
	if err != nil {
		log.Fatalln(err)
	}

	// TODO: check for auth key
	mux.HandleFunc("POST /api/links", func(w http.ResponseWriter, r *http.Request) {
		authKey := r.Header.Get("Authorization")
		if authKey != apiKey {
			WriteError(w, http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			WriteError(w, http.StatusInternalServerError)
		}

		payload := CreateLinkPayload{}

		if err = json.Unmarshal(body, &payload); err != nil {
			WriteError(w, http.StatusBadRequest)
			log.Println(err)
			return
		}

		if err = validate.Struct(payload); err != nil {
			WriteError(w, http.StatusBadRequest)
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
				WriteError(w, http.StatusNotFound)
				return
			}

			log.Println(err)
			WriteError(w, http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, url.Url, http.StatusMovedPermanently)
	})

	server := http.Server{
		Addr:    host,
		Handler: mux,
	}

	log.Printf("Server is starting: %s", host)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
