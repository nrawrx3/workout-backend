package model

import (
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
	return WorkoutPushups, constants.ErrWrongEnumString
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
	UserName string
	Email    string
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
