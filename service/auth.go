package service

import (
	"context"
	"errors"
	"reminder/auth"
	apperrors "reminder/errors"
	"reminder/models"

	"strconv"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (string, error)
	Register(ctx context.Context, username, password string) (models.User, error)
}

type authService struct {
	db     *gorm.DB
	tokens *auth.TokenService
}

func NewAuthService(db *gorm.DB, tokens *auth.TokenService) AuthService {
	return &authService{db: db, tokens: tokens}
}

func (a *authService) Login(ctx context.Context, username, password string) (string, error) {
	var user models.User
	if err := a.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", apperrors.ErrUnauthorized
		}
		return "", apperrors.ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", apperrors.ErrUnauthorized
	}
	return a.tokens.Issue(strconv.FormatUint(uint64(user.ID), 10))
}

func (a *authService) Register(ctx context.Context, username, password string) (models.User, error) {
	var user models.User
	if err := a.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, apperrors.ErrInternal
	}
	if user.ID > 0 {
		return models.User{}, apperrors.ErrConflict
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, apperrors.ErrInternal
	}
	user.Username = username
	user.PasswordHash = string(hash)
	if err := a.db.Create(&user).Error; err != nil {
		return models.User{}, apperrors.ErrInternal
	}
	return user, nil
}
