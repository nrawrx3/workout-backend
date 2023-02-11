package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nrawrx3/workout-backend/constants"
	"github.com/nrawrx3/workout-backend/model"
	"github.com/nrawrx3/workout-backend/store"
	"github.com/nrawrx3/workout-backend/util"
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
		cookieValueRaw, err := util.ReadCookieDecodeB64ThenDecrypt(r, h.sessionInfo.CookieName, h.cipher)

		sendResponse := func(errorMessage string) {
			if h.RedirectOnInvalidCookie {
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
			if !errors.Is(err, constants.ErrCodeNotFound) {
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

		uintSessionID, err := util.Uint64FromStringID(cookieValue.SessionID)
		if err != nil {
			errorMessage = "Invalid cookie data, failed to parse decrypted JSON"
			log.Printf("Invalid cookie data, failed to parse JSON: %v", err)
			sendResponse(errorMessage)
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
