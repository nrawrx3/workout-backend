package middleware

import (
	"errors"
	"net/http"

	"github.com/nrawrx3/workout-backend/constants"
	"github.com/rs/zerolog/log"
)

func Recover(wrapped http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				var err error
				switch t := rec.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = constants.ErrCodeUnknown
				}

				http.Error(w, "internal-server-error", http.StatusInternalServerError)

				log.Error().Err(err).Str("path", r.URL.Path).Str("remote-address", r.RemoteAddr).Str("forwarded-for", r.Header.Get("X-Forwarded-For")).Msg("recovered from panic in handler")
			}
		}()
		wrapped.ServeHTTP(w, r)
	})
}
