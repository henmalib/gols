package handlers

import (
	"fmt"
	"net/http"

	"github.com/henmalib/gols/packages/web/env"
	"github.com/henmalib/gols/packages/web/helpers"
)

// TODO: do something like auth middleware
func (state *App) DeleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	authKey := r.Header.Get("Authorization")
	if authKey != env.Env.ApiKey {
		helpers.WriteError(w, http.StatusUnauthorized)
		return
	}

	shortLink := r.URL.Query().Get("short")
	if shortLink == "" {
		http.Error(w, "You can't pass empty short link", http.StatusBadRequest)
		return
	}

	result, err := state.DB.Exec("DELETE FROM urls WHERE short = ?", shortLink)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	amount, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "RowsAffected: %d", amount)
}
