package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nrawrx3/workout-backend/constants"
	"github.com/nrawrx3/workout-backend/model"
	"github.com/nrawrx3/workout-backend/store"
	"github.com/nrawrx3/workout-backend/util"
)

type LoginHandler struct {
	userStore *store.UserStore
}

func NewLoginHandler(userStore *store.UserStore) *LoginHandler {
	return &LoginHandler{userStore: userStore}
}

func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	var reqBody model.UserLoginRequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse request"))
		return
	}
	defer r.Body.Close()
	user, err := h.userStore.GetUserWithEmail(reqBody.Email)
	if errors.Is(err, constants.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No user with given email"))
		return
	}

	passwordMatches, err := util.PasswordMatchesHash(reqBody.Password, user.PasswordHash)
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

	// Set cookie
}

// func (h *LoginHandler) AmILoggedIn(w http.ResponseWriter, r *http.Request) {
// 	var reqBody model.AmILoggedInRequestBody

// }
