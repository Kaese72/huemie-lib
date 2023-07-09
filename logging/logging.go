package logging

import (
	"fmt"
	"runtime"

	"github.com/Kaese72/huemie-lib/liberrors"
	"github.com/pkg/errors"
)

type Logger interface {
	Log(string, int, ...map[string]interface{})
}

var logger Logger = JSONLogger{}

func SetLogger(newLogger Logger) {
	logger = newLogger
}

var debugLogging bool = false

func Debug(msg string, data ...map[string]interface{}) {
	if debugLogging {
		logger.Log(msg, 7, data...)
	}
}

func Info(msg string, data ...map[string]interface{}) {
	logger.Log(msg, 6, data...)
}

func Warn(msg string, data ...map[string]interface{}) {
	logger.Log(msg, 4, data...)
}

func Error(msg string, data ...map[string]interface{}) {
	logger.Log(msg, 3, data...)
}

func Fatal(msg string, data ...map[string]interface{}) {
	logger.Log(msg, 1, data...)
}

func extractErrInfo(err error) map[string]interface{} {
	labels := map[string]interface{}{}
	earlyTracer := liberrors.EarliestTracer(err)
	if earlyTracer != nil {
		labels["trace.early"] = fmt.Sprintf("%+v", earlyTracer.StackTrace())
	}
	err = errors.Cause(err)
	if err != nil {
		labels["cause.error"] = err.Error()
	}
	return labels
}

func ErrError(err error, data ...map[string]interface{}) {
	logger.Log(err.Error(), 3, append(data, extractErrInfo(err))...)
}

func ErrWarn(err error, data ...map[string]interface{}) {
	logger.Log(err.Error(), 4, append(data, extractErrInfo(err))...)
}

func SetDebugLogging(flag bool) {
	debugLogging = flag
}

func mergeMaps(datas ...map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for _, data := range datas {
		for key, value := range data {
			merged[key] = value
		}
	}
	return merged
}

func collectData() map[string]interface{} {
	_, file, no, _ := runtime.Caller(3)
	return map[string]interface{}{
		"FILE": file,
		"LINE": no,
	}
}
