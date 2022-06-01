package xhttp

import (
	"encoding/json"
	"github.com/OdysseyMomentumExperience/token-service/pkg/log"
	"net/http"
	"strconv"
)

type APIError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
} // @name Error

func Error(w http.ResponseWriter, err error, code int) bool {
	if err == nil {
		return false
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)

	// Depending on the error code, log it as either an Error or DebugF
	if code == http.StatusInternalServerError {
		log.Error(err)
		err = enc.Encode(APIError{Message: "Something went wrong", Code: strconv.Itoa(code)})
	} else {
		log.Logln(0, err)
		err = enc.Encode(APIError{Message: err.Error(), Code: strconv.Itoa(code)})
	}

	if err != nil {
		panic(err) // If this happens, it's a programmer mistake so we panic
	}

	return true
}
