package model

import backend_model "github.com/nrawrx3/workout-backend/model"

func (w WorkoutKind) CastToModelKind() backend_model.WorkoutKind {
	switch w {
	case WorkoutKindPushUps:
		return backend_model.WorkoutPushups
	case WorkoutKindKneesOverToes:
		return backend_model.WorkoutKneesOverToes
	case WorkoutKindBurpees:
		return backend_model.WorkoutBurpees
	case WorkoutKindOneTwos:
		return backend_model.WorkoutOneTwos
	default:
		return backend_model.WorkoutPushups
	}
}

func WorkoutKindFromModel(kind backend_model.WorkoutKind) WorkoutKind {
	switch kind {
	case backend_model.WorkoutPushups:
		return WorkoutKindPushUps
	case backend_model.WorkoutKneesOverToes:
		return WorkoutKindKneesOverToes
	case backend_model.WorkoutBurpees:
		return WorkoutKindBurpees
	case backend_model.WorkoutOneTwos:
		return WorkoutKindOneTwos
	default:
		return WorkoutKindPushUps
	}
}
