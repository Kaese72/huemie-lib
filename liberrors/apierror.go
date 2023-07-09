package liberrors

import "net/http"

type ApiReason int

const (
	NotFound      ApiReason = http.StatusNotFound
	UserError     ApiReason = http.StatusBadRequest
	InternalError ApiReason = http.StatusInternalServerError
)

type apiError struct {
	Reason ApiReason
	err    error
}

func (err apiError) Error() string {
	return err.err.Error()
}

func (err apiError) Unwrap() error {
	return err.err
}

func (err apiError) Cause() error {
	return err.err
}

func ApiError(reason ApiReason, err error) error {
	if err == nil {
		return nil
	}
	return apiError{
		Reason: reason,
		err:    err,
	}
}
