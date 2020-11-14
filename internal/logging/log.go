package logging

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"os"
)

// NewLogger creates a filtered log
func NewLogger(logLevel string, appIdentifier string) *log.Logger {
	var logger log.Logger
	logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	logger = setLevelKey(logger, "severity")
	switch logLevel {
	case "debug":
		logger = level.NewFilter(logger, level.AllowDebug())
	case "info":
		logger = level.NewFilter(logger, level.AllowInfo())
	case "warn":
		logger = level.NewFilter(logger, level.AllowWarn())
	case "error":
		logger = level.NewFilter(logger, level.AllowError())
	default:
		logger = level.NewFilter(logger, level.AllowAll())
	}
	logger = log.With(logger, "chronicler", appIdentifier)
	return &logger
}

func setLevelKey(logger log.Logger, key interface{}) log.Logger {
	return log.LoggerFunc(func(keyvals ...interface{}) error {
		for i := 1; i < len(keyvals); i += 2 {
			if _, ok := keyvals[i].(level.Value); ok {
				// overwriting the key without copying keyvals
				// techically violates the log.Logger contract
				// but is safe in this context because none
				// of the loggers in this program retain a reference
				// to keyvals
				keyvals[i-1] = key
				break
			}
		}
		return logger.Log(keyvals...)
	})
}