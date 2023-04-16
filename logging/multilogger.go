package logging

type MultiLogger struct {
	Loggers []Logger
}

func (logger MultiLogger) Log(msg string, priority int, datas ...map[string]interface{}) {
	for _, sublogger := range logger.Loggers {
		sublogger.Log(msg, priority, datas...)
	}
}
