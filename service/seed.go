package service

import (
	"reminder/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUser(db *gorm.DB, username, password string) error {
	var count int64
	db.Model(&models.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return db.Create(&models.User{Username: username, PasswordHash: string(hash)}).Error
}
