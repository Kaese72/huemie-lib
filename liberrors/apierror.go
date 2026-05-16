package liberrors

import (
	"encoding/json"
	"net/http"
)

type ApiReason int

const (
	NotFound      ApiReason = http.StatusNotFound
	UserError     ApiReason = http.StatusBadRequest
	Unauthorized  ApiReason = http.StatusUnauthorized
	InternalError ApiReason = http.StatusInternalServerError
)

type ApiError struct {
	Reason ApiReason
	Err    error
}

func (err ApiError) Error() string {
	return err.Err.Error()
}

func (err ApiError) Unwrap() error {
	return err.Err
}

func (err ApiError) Cause() error {
	return err.Err
}

// WriteHTTP writes the error as a JSON response to w.
func (err ApiError) WriteHTTP(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(err.Reason))
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: err.Err.Error()})
}

func NewApiError(reason ApiReason, err error) *ApiError {
	if err == nil {
		return nil
	}
	return &ApiError{
		Reason: reason,
		Err:    err,
	}
}
