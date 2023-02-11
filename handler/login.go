package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/nrawrx3/workout-backend/constants"
	"github.com/nrawrx3/workout-backend/model"
	"github.com/nrawrx3/workout-backend/store"
	"github.com/nrawrx3/workout-backend/util"
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

	log.Printf("received /login with form-data: %+v", formData)

	user, err := h.userStore.GetUserWithEmail(r.Context(), formData.Email)
	if errors.Is(err, constants.ErrCodeNotFound) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No user with given email"))
		return
	}

	passwordMatches, err := util.PasswordMatchesHash(formData.Password, user.PasswordHash)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("failed to match password due to error"))
		return
	}
	if !passwordMatches {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Wrong password"))
		return
	}

	// Create session
	session, err := h.userStore.CreateSession(r.Context(), user.ID, time.Now(), h.cookieInfo.Expires, r.Header.Get("User-Agent"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.DefaultInternalServerErrorResponse)
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
		log.Printf("failed to JSON-encode cookie value: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to encode cookie :|"))
		return
	}

	err = util.EncryptThenEncodeB64ThenWriteCookie(w, cookie, h.cipher, cookieValueBuf.Bytes())
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("fauled to encrypt cookie :|"))
		return
	}
	w.WriteHeader(http.StatusOK)
	log.Printf("successfully logged in user: %d", user.ID)
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

	sessionId, err := util.ExtractSessionIDFromCookie(r, h.cookieInfo.CookieName, h.cipher)
	if err != nil {
		log.Printf("could not extract cookie value: %v", err)
		response.ErrorCode = constants.ResponseErrCodeUserNotLoggedIn
	} else {
		session, err := h.userStore.LoadSession(r.Context(), sessionId, time.Now())
		if errors.Is(err, constants.ErrCodeNotFound) {
			response.ErrorCode = constants.ResponseErrCodeUnexpectedServerError
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.ErrorCode = constants.ResponseErrCodeUserNotLoggedIn
		} else {
			log.Printf("AmILoggedIn called: User %d is logged in already to session %d", session.UserID, sessionId)
			response.Data = model.AmILoggedInResponseJSON{LoggedIn: true}
		}
	}

	util.AddJsonContentHeader(w, 0)
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to encode response json: %v", err)
		return
	}
}
