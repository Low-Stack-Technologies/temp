package http_error

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) ToJSON() string {
	json, _ := json.Marshal(e)
	return string(json)
}

func (e *Error) FromJSON(str string) error {
	return json.Unmarshal([]byte(str), e)
}

func (e *Error) Respond(w http.ResponseWriter) {
	w.WriteHeader(e.StatusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(e.ToJSON()))
}

func Respond(w http.ResponseWriter, statusCode int, message string) {
	e := Error{StatusCode: statusCode, Message: message}
	e.Respond(w)
}
