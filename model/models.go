package model

import (
	"net/http"
	"strconv"
	"time"

	"github.com/nrawrx3/workout-backend/constants"
	"gorm.io/gorm"
)

type WorkoutKind string

const (
	WorkoutPushups       WorkoutKind = "pushups"
	WorkoutOneTwos       WorkoutKind = "onetwos"
	WorkoutBurpees       WorkoutKind = "burpees"
	WorkoutKneesOverToes WorkoutKind = "kneesovertoes"
)

func CastWorkoutKind(str string) (WorkoutKind, error) {
	switch str {
	case string(WorkoutPushups):
		return WorkoutPushups, nil
	case string(WorkoutBurpees):
		return WorkoutBurpees, nil
	case string(WorkoutKneesOverToes):
		return WorkoutKneesOverToes, nil
	case string(WorkoutOneTwos):
		return WorkoutOneTwos, nil
	}
	return WorkoutPushups, constants.ErrCodeWrongEnumString
}

type BaseModel struct {
	ID        uint64 `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Object model corresponding to users table
type User struct {
	BaseModel
	UserName     string
	Email        string
	PasswordHash string
}

type UserSession struct {
	BaseModel
	ExpiresAt time.Time
	UserAgent string
	UserID    uint64
	User      User
}

// Object model corresponding to workouts table
type Workout struct {
	BaseModel
	Kind            WorkoutKind
	Reps            int
	Rounds          int
	DurationSeconds int
	Order           int `gorm:"column:relative_order"`
	UserID          uint64
	User            User
}

type UserLoginRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AmILoggedInRequestBody struct {
	Extra string `json:"extra"`
}

type WorkoutResponseJSON struct {
	ID              string `json:"id"`
	Kind            string `json:"kind"`
	Reps            int    `json:"reps"`
	DurationSeconds int    `json:"duration_seconds"`
	Order           int    `json:"order"`
	UserID          string `json:"user_id"`
}

func (resp *WorkoutResponseJSON) FromModel(w *Workout) {
	resp.ID = strconv.FormatUint(w.ID, 10)
	resp.Kind = string(w.Kind)
	resp.Reps = w.Reps
	resp.DurationSeconds = w.DurationSeconds
	resp.Order = w.Order
	resp.UserID = strconv.FormatUint(w.UserID, 10)
}

type WorkoutListResponseJSON struct {
	Workouts []WorkoutResponseJSON `json:"workouts"`
}

type ResponseFormatJSON struct {
	Data         interface{} `json:"data"`
	ErrorCode    string      `json:"error_code"`
	ErrorMessage string      `json:"error_message"`
}

// Doesn't make sense to send a JSON response in cases of internal server error.
// What if json.Encode(...) itself fails to write to the response.
var DefaultInternalServerErrorResponse = ResponseFormatJSON{
	Data:         nil,
	ErrorCode:    constants.ResponseErrCodeUnexpectedServerError,
	ErrorMessage: "unexpected server side error",
}

var UserNotLoggedInErrorResponse = ResponseFormatJSON{
	Data:         nil,
	ErrorCode:    constants.ResponseErrCodeUserNotLoggedIn,
	ErrorMessage: "user not logged in",
}

// key type for the request context value containing the user id extracted from
// cookie
type UserIDContextKey struct{}

// key type for the request context value containing the UserSession object
type UserSessionContextKey struct{}

type AmILoggedInResponseJSON struct {
	LoggedIn bool `json:"logged_in"`
}

type SessionCookieInfo struct {
	CookieName string
	Secure     bool
	SecretKey  string
	Domain     string
	SameSite   http.SameSite
	Expires    time.Time
	HttpOnly   bool
}

type SessionCookieValue struct {
	SessionID string `json:"session_id"`
}
