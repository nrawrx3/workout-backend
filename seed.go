package backend

import (
	"github.com/nrawrx3/workout-backend/model"
	"github.com/nrawrx3/workout-backend/util"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func SeedDatabase(db *gorm.DB) error {
	const password = "sigmamale"
	passwordHash, err := util.HashPasswordBase64(password)

	if err != nil {
		return err
	}

	user := model.User{
		UserName:     "jane",
		Email:        "jane@example.com",
		PasswordHash: passwordHash,
	}

	err = db.Create(&user).Error
	if err != nil {
		return errors.WithMessage(err, "failed to create user")
	}

	workouts := []model.Workout{
		{
			Kind:            model.WorkoutPushups,
			Reps:            10,
			Rounds:          2,
			DurationSeconds: 20,
			UserID:          user.ID,
			Order:           0,
		},
		{
			Kind:            model.WorkoutOneTwos,
			Reps:            100,
			Rounds:          1,
			DurationSeconds: 60,
			UserID:          user.ID,
			Order:           1,
		},
		{
			Kind:            model.WorkoutKneesOverToes,
			Reps:            20,
			Rounds:          1,
			DurationSeconds: 20,
			UserID:          user.ID,
			Order:           2,
		},
		{
			Kind:            model.WorkoutBurpees,
			Reps:            20,
			Rounds:          1,
			DurationSeconds: 60,
			UserID:          user.ID,
			Order:           3,
		},
	}

	for i := range workouts {
		wk := &workouts[i]
		err := db.Create(wk).Error
		if err != nil {
			return errors.Wrapf(err, "failed to create workout %+v", wk)
		}
	}
	return nil
}
