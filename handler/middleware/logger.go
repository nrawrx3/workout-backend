package middleware

import (
	"net/http"

	"github.com/felixge/httpsnoop"
	"github.com/rs/zerolog/log"
)

func Logger(wrapped http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(wrapped, w, r)
		log.Info().Str("method", r.Method).Str("path", r.URL.Path).Str("host", r.URL.Host).Int("status-code", m.Code).Int64("duration", m.Duration.Milliseconds()).Int64("bytes", m.Written).Msg("http-handler")
	})
}
