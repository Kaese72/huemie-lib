package liberrors

import (
	"net/http"
)

type ApiReason int

const (
	NotFound      ApiReason = http.StatusNotFound
	UserError     ApiReason = http.StatusBadRequest
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

func NewApiError(reason ApiReason, err error) *ApiError {
	if err == nil {
		return nil
	}
	return &ApiError{
		Reason: reason,
		Err:    err,
	}
}
