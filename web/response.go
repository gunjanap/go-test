package web

import (
	"encoding/json"
	"net/http"

	microlog "github.com/islax/microapp/log"
)

// RespondJSON makes the response with payload as json format
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		microlog.Logger.Error([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

// RespondErrorMessage makes the error response with payload as json format
func RespondErrorMessage(w http.ResponseWriter, code int, message string) {
	RespondJSON(w, code, map[string]string{"error": message})
}

// RespondError returns a validation error else
func RespondError(w http.ResponseWriter, err error) {
	switch err.(type) {
	case ValidationError:
		RespondJSON(w, http.StatusBadRequest, err)
	case HTTPError:
		httpError := err.(HTTPError)
		RespondErrorMessage(w, httpError.HTTPStatus, httpError.ErrorKey)
	default:
		RespondErrorMessage(w, http.StatusInternalServerError, "Key_InternalError")
	}
}
