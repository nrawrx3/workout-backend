package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/nrawrx3/workout-backend/constants"
	"github.com/nrawrx3/workout-backend/model"
	"github.com/nrawrx3/workout-backend/util"
)

type SessionRedirectToLogin struct {
	sessionInfo             model.SessionCookieInfo
	cipher                  *util.AESCipher
	RedirectOnInvalidCookie bool
}

func NewSessionRedirectToLogin(sessionInfo model.SessionCookieInfo, cipher *util.AESCipher) *SessionRedirectToLogin {
	return &SessionRedirectToLogin{sessionInfo: sessionInfo, cipher: cipher}
}

func (s *SessionRedirectToLogin) Handler(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Fetch cookie value
		cookieValueRaw, err := util.ReadCookieDecodeB64ThenDecrypt(r, s.sessionInfo.CookieName, s.cipher)

		sendResponse := func(errorMessage string) {
			if s.RedirectOnInvalidCookie {
				log.Printf("redirecting request from %s to /login", r.RemoteAddr)
				http.Redirect(w, r, constants.LoginPath, http.StatusSeeOther)
			} else {
				log.Printf("sending 401 Bad Request response %s because expected session cookie was not received", r.RemoteAddr)

				responseData := model.UserNotLoggedInErrorResponse
				responseData.ErrorMessage = errorMessage

				util.AddJsonContentHeader(w, http.StatusUnauthorized)
				if err := json.NewEncoder(w).Encode(&model.UserNotLoggedInErrorResponse); err != nil {
					log.Printf("unexpected json encoding error: %v", err)
				}
			}
		}

		errorMessage := "cookie unset or expired"

		if err != nil {
			if !errors.Is(err, constants.ErrNotFound) {
				errorMessage = "failed to decode cookie"
			}
			sendResponse(errorMessage)
			return
		}

		var cookieValue model.SessionCookieValue
		err = json.NewDecoder(strings.NewReader(cookieValueRaw)).Decode(&cookieValue)
		if err != nil {
			errorMessage = "Invalid cookie data, failed to parse decrypted JSON"
			log.Printf("Invalid cookie data, failed to parse JSON: %v", err)
			sendResponse(errorMessage)
			return
		}

		uintUserID, err := util.Uint64FromStringID(cookieValue.UserID)
		if err != nil {
			errorMessage = "Invalid cookie data, failed to parse decrypted JSON"
			log.Printf("Invalid cookie data, failed to parse JSON: %v", err)
			sendResponse(errorMessage)
			return
		}

		ctx := context.WithValue(r.Context(), model.UserIDContextKey{}, uintUserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
