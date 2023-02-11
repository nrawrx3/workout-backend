package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/nrawrx3/workout-backend/model"
	"github.com/nrawrx3/workout-backend/store"
	"github.com/nrawrx3/workout-backend/util"
)

type WorkoutsListHandler struct {
	userStore *store.UserStore
}

func NewWorkoutsListHandler(userStore *store.UserStore) *WorkoutsListHandler {
	return &WorkoutsListHandler{userStore: userStore}
}

func (h *WorkoutsListHandler) HandleGetWorkoutsList(w http.ResponseWriter, r *http.Request) {
	session, ok := r.Context().Value(model.UserSessionContextKey{}).(model.UserSession)
	if !ok {
		log.Printf("No model.UserIDContextKey{} in context values map!")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&model.DefaultInternalServerErrorResponse)
		return
	}

	workouts, err := h.userStore.GetWorkoutsOfUser(r.Context(), session.UserID)
	if err != nil {
		log.Printf("failed to retrieve workout list of user: %d: %v", session.UserID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&model.DefaultInternalServerErrorResponse)
		return
	}

	resp := model.ResponseFormatJSON{}
	workoutListJSON := model.WorkoutListResponseJSON{
		Workouts: make([]model.WorkoutResponseJSON, 0, len(workouts)),
	}

	for _, w := range workouts {
		var workoutJSON model.WorkoutResponseJSON
		workoutJSON.FromModel(&w)
		workoutListJSON.Workouts = append(workoutListJSON.Workouts, workoutJSON)
	}
	resp.Data = workoutListJSON
	util.AddJsonContentHeader(w, http.StatusOK)
	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		log.Print(err)
		util.AddJsonContentHeader(w, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&model.DefaultInternalServerErrorResponse)
		return
	}
}
