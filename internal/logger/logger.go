package logger

import (
	"charm.land/log/v2"
	"io"
)

type Logger interface {
	Debug(msg any, keyvals ...any)
	Info(msg any, keyvals ...any)
	Warn(msg any, keyvals ...any)
	Error(msg any, keyvals ...any)
	Print(msg any, keyvals ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Printf(msg string, keyvals ...any)
}

func New(w io.Writer, level string, fmt string) Logger {
	opts := log.Options{
		ReportCaller: true,
		Prefix:       "veritas ⚖️",
	}

	switch level {
	case "debug":
		opts.Level = log.DebugLevel
	case "info":
		opts.Level = log.InfoLevel
	case "warn":
		opts.Level = log.WarnLevel
	case "error":
		opts.Level = log.ErrorLevel
	default:
		opts.Level = log.InfoLevel
	}

	switch fmt {
	case "text":
		opts.Formatter = log.TextFormatter
	case "json":
		opts.Formatter = log.JSONFormatter
	default:
		opts.Formatter = log.TextFormatter
	}

	return log.NewWithOptions(w, opts)
}
