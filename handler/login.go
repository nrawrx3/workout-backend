package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/nrawrx3/uno-backend/constants"
	"github.com/nrawrx3/uno-backend/model"
	"github.com/nrawrx3/uno-backend/store"
	"github.com/nrawrx3/uno-backend/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LoginHandler struct {
	userStore  *store.UserStore
	cookieInfo model.SessionCookieInfo
	cipher     *util.AESCipher
}

func NewLoginHandler(userStore *store.UserStore, cookieInfo model.SessionCookieInfo, cipher *util.AESCipher) *LoginHandler {
	return &LoginHandler{userStore: userStore, cookieInfo: cookieInfo, cipher: cipher}
}

// Success response type: 200 - empty
// Failure response type:
//
//	404 - reason-string
//	422 - reson-string
//	500 - model.DefaultInternalServerErrorResponse
func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	<-time.After(2 * time.Second)
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("expected POST method to API"))
		return
	}

	formData := model.UserLoginRequestBody{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}

	user, err := h.userStore.GetUserWithEmail(r.Context(), formData.Email)
	if errors.Is(err, constants.ErrCodeNotFound) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No user with given email"))
		return
	}

	passwordMatches, err := util.PasswordMatchesHash(formData.Password, user.PasswordHash)

	if err != nil {
		log.Debug().Str("/login", "password does not match, non-trivial error").Err(err).Send()
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("failed to match password due to error"))
		return
	}
	if !passwordMatches {
		log.Debug().Str("/login", "password does not match")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Wrong password"))
		return
	}

	// Create session
	session, err := h.userStore.CreateSession(r.Context(), user.ID, time.Now(), h.cookieInfo.Expires, r.Header.Get("User-Agent"))
	if err != nil {
		http.Error(w, constants.ResponseErrCodeUnexpectedServerError, http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name: h.cookieInfo.CookieName,
		// Value:    strconv.FormatUint(user.ID, 10),
		Domain:   h.cookieInfo.Domain,
		Expires:  h.cookieInfo.Expires,
		HttpOnly: h.cookieInfo.HttpOnly,
		SameSite: h.cookieInfo.SameSite,
		Secure:   h.cookieInfo.Secure,
	}

	cookieValue := model.SessionCookieValue{
		SessionID: strconv.FormatUint(session.ID, 10),
	}
	cookieValueBuf := bytes.NewBuffer(nil)
	err = json.NewEncoder(cookieValueBuf).Encode(&cookieValue)
	if err != nil {
		log.Error().Str("path", "/login").Err(err).Msg("failed to JSON encode cookie value")
		http.Error(w, "failed to JSON encode cookie :|", http.StatusInternalServerError)
		return
	}

	err = model.EncryptThenEncodeB64ThenWriteCookie(w, cookie, h.cipher, cookieValueBuf.Bytes())
	if err != nil {
		log.Error().Str("path", "/login").Err(err).Msg("failed to write cookie")
		http.Error(w, "failed to encrypt cookie :|", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	log.Info().Str("/login", "logged in user").Uint64("userID", user.ID).Send()
}

// Success response type: 200 - model.AmILoggedInResponseJSON
func (h *LoginHandler) AmILoggedIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("expected a GET request, received different"))
		return
	}

	w.WriteHeader(http.StatusOK)

	response := model.ResponseFormatJSON{}

	sessionId, err := model.ExtractSessionIDFromCookie(r, h.cookieInfo.CookieName, h.cipher)
	if err != nil {
		log.Info().Err(err).Str("path", "/am-i-logged-in").Msg("could not extract sessionID from cookie")
		response.ErrorCode = constants.ResponseErrCodeUserNotLoggedIn
	} else {
		session, err := h.userStore.LoadSession(r.Context(), sessionId, time.Now())
		if errors.Is(err, constants.ErrCodeNotFound) {
			response.ErrorCode = constants.ResponseErrCodeUnexpectedServerError
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.ErrorCode = constants.ResponseErrCodeUserNotLoggedIn
		} else {
			log.Info().Str("/am-i-logged-in", "user is logged in").Dict("session", zerolog.Dict().Uint64("sessionID", sessionId).Uint64("userID", session.UserID)).Send()
			response.Data = model.AmILoggedInResponseJSON{LoggedIn: true}
		}
	}

	util.AddJsonContentHeader(w, 0)
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		http.Error(w, "failed to encode response json", http.StatusInternalServerError)
		log.Info().Str("/am-i-logged-in", "failed to encode response json").Err(err).Send()
		return
	}
}
