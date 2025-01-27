package helpers

import "net/http"

func WriteError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
