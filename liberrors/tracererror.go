package liberrors

import (
	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func EarliestTracer(err error) stackTracer {
	if err == nil {
		return nil
	}
	var newRoot error = err
	for {
		var tracer stackTracer
		if errors.As(newRoot, &tracer) {
			if insiderErr := errors.Unwrap(newRoot); insiderErr != nil {
				newRoot = insiderErr
			} else {
				return tracer
			}
		} else {
			return nil
		}
	}
}
