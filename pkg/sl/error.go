package sl

import "log/slog"

// Error wraps errors as slog attribute
func Error(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
