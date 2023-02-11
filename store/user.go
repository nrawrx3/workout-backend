package store

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

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

func (s *UserStore) GetUserWithEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := s.DB.Where("email = ?", []interface{}{email}).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user, constants.ErrCodeNotFound
		}
		return user, fmt.Errorf("failed to find user with email: %s: %w", email, err)
	}
	return user, nil
}

func (s *UserStore) GetWorkoutsOfUser(ctx context.Context, userId uint64) ([]model.Workout, error) {
	var workouts []model.Workout
	err := s.DB.Where("user_id = ?", userId).Find(&workouts).Error
	if err != nil {
		return nil, err
	}
	return workouts, nil
}

// Gets the model.UserSession corresponding to given session id if it exists and
// not expired in the user_sessions table. Preloads the User field.
func (s *UserStore) LoadSession(ctx context.Context, sessionId uint64, timeNow time.Time) (model.UserSession, error) {
	session := model.UserSession{
		BaseModel: model.BaseModel{ID: sessionId},
	}
	err := s.DB.WithContext(ctx).Preload("User").Where("expires_at > ?", timeNow).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return session, constants.ErrCodeNotFound
		}

		log.Printf("error when fetching sessionId %d: %v", sessionId, err)
		return session, err
	}

	return session, nil
}

func (s *UserStore) CreateSession(ctx context.Context, userId uint64, timeNow, expiresAt time.Time, userAgent string) (model.UserSession, error) {
	// log.Printf("CreateSession called")
	session := model.UserSession{
		UserID: userId,
	}

	// Check if session already exists and not-expired. If yes, send that.
	// This makes CreateSession idempotent.
	tx := s.DB.Begin()
	defer tx.Rollback()

	err := tx.WithContext(ctx).Preload("User").Where("user_id = ? and expires_at > ? and user_agent = ?", userId, timeNow, userAgent).First(&session).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("query error: %v", err)
			return session, err
		}
		err = nil
	} else {
		log.Printf("returning existing session (id = %d) for userId (id = %d)", session.ID, session.User.ID)
		return session, nil
	}

	err = tx.Model(model.UserSession{}).Where("user_agent = ?", userAgent).Error
	if err != nil {
		log.Printf("CreateSession failed to delete sessions for userId %d with same user agent: %v", userId, err)
		return model.UserSession{}, err
	}

	// Create
	session = model.UserSession{
		UserID:    userId,
		ExpiresAt: expiresAt,
		UserAgent: userAgent,
	}

	err = tx.Create(&session).Error
	if err != nil {
		log.Printf("failed to create session in store: %v", err)
		return session, err
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("CreateSession failed to commit: %v", err)
		return session, err
	}
	return session, nil
}
