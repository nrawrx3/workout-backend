package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

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

func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("expected POST method to API"))
		return
	}

	// var reqBody model.UserLoginRequestBody
	// err := json.NewDecoder(r.Body).Decode(&reqBody)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	w.Write([]byte("Failed to parse request"))
	// 	return
	// }
	// defer r.Body.Close()
	formData := model.UserLoginRequestBody{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}

	log.Printf("received /login with form-data: %+v", formData)

	user, err := h.userStore.GetUserWithEmail(formData.Email)
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
		UserID: strconv.FormatUint(user.ID, 10),
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

func (h *LoginHandler) AmILoggedIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("expected a GET request, received different"))
		return
	}

	w.WriteHeader(http.StatusOK)

	response := model.ResponseFormatJSON{}

	userId, err := util.ReadCookieDecodeB64ThenDecrypt(r, h.cookieInfo.CookieName, h.cipher)
	if err != nil {
		response.Data = model.AmILoggedInResponseJSON{
			LoggedIn: false,
		}

		if errors.Is(err, constants.ErrCodeNotFound) {
			response.ErrorCode = constants.ErrCodeNotFound.Error()
		} else {
			response.ErrorCode = constants.ErrCodeUnknown.Error()
		}

		err := json.NewEncoder(w).Encode(&response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("failed to encode response json: %v", err)
			return
		}
		return
	}

	log.Printf("AmILoggedIn called: User %s is logged in already", userId)
	response.Data = model.AmILoggedInResponseJSON{LoggedIn: true}
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to encode response json: %v", err)
		return
	}
}
