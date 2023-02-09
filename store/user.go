package store

import (
	"errors"
	"fmt"

	"github.com/nrawrx3/workout-backend/constants"
	"github.com/nrawrx3/workout-backend/model"
	"gorm.io/gorm"
)

type UserStore struct {
	DB *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{DB: db}
}

func (s *UserStore) GetUserWithEmail(email string) (model.User, error) {
	var user model.User
	err := s.DB.Model(&user).Where("email = ?", []interface{}{email}).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user, constants.ErrCodeNotFound
		}
		return user, fmt.Errorf("failed to find user with email: %s: %w", email, err)
	}
	return user, nil
}

func (s *UserStore) GetWorkoutsOfUser(userId uint64) ([]model.Workout, error) {
	var workouts []model.Workout
	err := s.DB.Where("user_id = ?", userId).Find(&workouts).Error
	if err != nil {
		return nil, err
	}
	return workouts, nil
}
