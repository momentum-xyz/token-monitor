package xhttp

import (
	"encoding/json"
	"net/http"
)

type APISuccess struct {
	Message string `json:"message"`
} // @name Success

func Success(w http.ResponseWriter, msg string) bool {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(200)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)

	err := enc.Encode(APISuccess{Message: msg})
	if err != nil {
		panic(err) // If this happens, it's a programmer mistake so we panic
	}

	return true
}
