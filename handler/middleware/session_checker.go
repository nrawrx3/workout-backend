package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/nrawrx3/uno-backend/constants"
	"github.com/nrawrx3/uno-backend/model"
	"github.com/nrawrx3/uno-backend/store"
	"github.com/nrawrx3/uno-backend/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type SessionChecker struct {
	sessionInfo             model.SessionCookieInfo
	cipher                  *util.AESCipher
	RedirectOnInvalidCookie bool
	userStore               *store.UserStore
}

func NewSessionChecker(userStore *store.UserStore, sessionInfo model.SessionCookieInfo, cipher *util.AESCipher) *SessionChecker {
	return &SessionChecker{sessionInfo: sessionInfo, cipher: cipher, userStore: userStore}
}

func (h *SessionChecker) Handler(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Fetch cookie value
		cookieValueRaw, err := model.ReadCookieDecodeB64ThenDecrypt(r, h.sessionInfo.CookieName, h.cipher)

		sendResponse := func(errorMessage string) {
			if h.RedirectOnInvalidCookie {
				log.Info().Dict("session-checker", zerolog.Dict().Str("remote-address", r.RemoteAddr)).Msg("redirecting to /login")
				http.Redirect(w, r, constants.LoginPath, http.StatusSeeOther)
			} else {
				log.Info().Dict("session-checker", zerolog.Dict().Str("remote-address", r.RemoteAddr).Str("request-path", r.URL.Path)).Msg("sending 401 Unauthorized")

				responseData := model.UserNotLoggedInErrorResponse
				responseData.ErrorMessage = errorMessage

				util.AddJsonContentHeader(w, http.StatusUnauthorized)
				if err := json.NewEncoder(w).Encode(&model.UserNotLoggedInErrorResponse); err != nil {
					log.Error().Err(err).Msg("unexpected json encoding error")
				}
			}
		}

		errorMessage := "cookie unset or expired"

		if err != nil {
			if !errors.Is(err, constants.ErrCodeNotFound) {
				errorMessage = "failed to decode cookie"
			}
			sendResponse(errorMessage)
			return
		}

		var cookieValue model.SessionCookieValue
		err = json.NewDecoder(strings.NewReader(cookieValueRaw)).Decode(&cookieValue)
		if err != nil {
			log.Info().Err(err).Msg("failed to parse JSON")
			sendResponse("Invalid cookie data, failed to parse decrypted JSON")
			return
		}

		uintSessionID, err := util.Uint64FromStringID(cookieValue.SessionID)
		if err != nil {
			log.Info().Err(err).Msg("failed to parse JSON, the SessionID could not be parsed as uint64")
			sendResponse("Invalid cookie data, failed to parse decrypted JSON")
			return
		}

		session, err := h.userStore.LoadSession(r.Context(), uintSessionID, time.Now())
		if err != nil {
			if errors.Is(err, constants.ErrCodeNotFound) {
				util.AddJsonContentHeader(w, http.StatusNotFound)
				json.NewEncoder(w).Encode(model.UserNotLoggedInErrorResponse)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(model.DefaultInternalServerErrorResponse)
			return
		}

		ctx := context.WithValue(r.Context(), model.UserSessionContextKey{}, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
