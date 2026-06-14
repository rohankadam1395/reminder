package models

import (
	"time"

	"gorm.io/gorm"
)

type Reminder struct {
	gorm.Model
	Message  string    `json:"message"`
	Time     time.Time `json:"time"`
	Notified bool      `json:"notified"`
	UserID   string    `json:"user_id"`
}
