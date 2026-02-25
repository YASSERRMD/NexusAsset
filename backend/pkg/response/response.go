package response

import (
	"encoding/json"
	"net/http"
)

// Meta holds pagination metadata for list responses.
type Meta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// Envelope is the standard API response wrapper.
//
// Example success:
//
//	{ "success": true, "data": {...}, "error": "", "meta": null }
//
// Example error:
//
//	{ "success": false, "data": null, "error": "record not found", "meta": null }
type Envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Error   string `json:"error"`
	Meta    *Meta  `json:"meta,omitempty"`
}

func write(w http.ResponseWriter, status int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// OK sends 200 with data.
func OK(w http.ResponseWriter, data any) {
	write(w, http.StatusOK, Envelope{Success: true, Data: data})
}

// OKList sends 200 with data and pagination meta.
func OKList(w http.ResponseWriter, data any, meta Meta) {
	write(w, http.StatusOK, Envelope{Success: true, Data: data, Meta: &meta})
}

// Created sends 201 with data.
func Created(w http.ResponseWriter, data any) {
	write(w, http.StatusCreated, Envelope{Success: true, Data: data})
}

// NoContent sends 204.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest sends 400 with an error message.
func BadRequest(w http.ResponseWriter, err string) {
	write(w, http.StatusBadRequest, Envelope{Success: false, Error: err})
}

// Unauthorized sends 401.
func Unauthorized(w http.ResponseWriter, err string) {
	write(w, http.StatusUnauthorized, Envelope{Success: false, Error: err})
}

// Forbidden sends 403.
func Forbidden(w http.ResponseWriter, err string) {
	write(w, http.StatusForbidden, Envelope{Success: false, Error: err})
}

// NotFound sends 404.
func NotFound(w http.ResponseWriter, err string) {
	write(w, http.StatusNotFound, Envelope{Success: false, Error: err})
}

// Conflict sends 409.
func Conflict(w http.ResponseWriter, err string) {
	write(w, http.StatusConflict, Envelope{Success: false, Error: err})
}

// UnprocessableEntity sends 422.
func UnprocessableEntity(w http.ResponseWriter, err string) {
	write(w, http.StatusUnprocessableEntity, Envelope{Success: false, Error: err})
}

// InternalError sends 500.
func InternalError(w http.ResponseWriter, err string) {
	write(w, http.StatusInternalServerError, Envelope{Success: false, Error: err})
}
