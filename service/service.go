package service

import (
	"context"
	"errors"
	"reminder/auth"
	apperrors "reminder/errors"
	"reminder/models"
	"time"

	"gorm.io/gorm"
)

type CursorReminders struct {
	Reminders []models.Reminder
	HasMore   bool
}

type ReminderService interface {
	SetReminder(ctx context.Context, reminder models.Reminder) (models.Reminder, error)
	GetReminders(ctx context.Context, afterID uint, limit int) (CursorReminders, error)
	GetReminder(ctx context.Context, id string) (models.Reminder, error)
	GetReminderByFilter(ctx context.Context, filter string) ([]models.Reminder, error)
	ListDueReminders(ctx context.Context, now time.Time, limit int) ([]models.Reminder, error)
	MarkReminderNotified(ctx context.Context, id uint) error
}

type reminderService struct {
	db *gorm.DB
}

func NewReminderService(db *gorm.DB) ReminderService {
	return &reminderService{db: db}
}

func userIDFromContext(ctx context.Context) (string, error) {
	id, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return "", apperrors.ErrUnauthorized
	}
	return id, nil
}

func (s *reminderService) SetReminder(ctx context.Context, reminder models.Reminder) (models.Reminder, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return models.Reminder{}, err
	}
	reminder.UserID = userID
	if err := s.db.Create(&reminder).Error; err != nil {
		return models.Reminder{}, err
	}
	return reminder, nil
}

func (s *reminderService) GetReminders(ctx context.Context, afterID uint, limit int) (CursorReminders, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return CursorReminders{}, err
	}
	query := s.db.Where("user_id = ?", userID).Order("id ASC").Limit(limit + 1)
	if afterID > 0 {
		query = query.Where("id > ?", afterID)
	}

	var reminders []models.Reminder
	if err := query.Find(&reminders).Error; err != nil {
		return CursorReminders{}, err
	}

	hasMore := len(reminders) > limit
	if hasMore {
		reminders = reminders[:limit]
	}
	return CursorReminders{Reminders: reminders, HasMore: hasMore}, nil
}

func (s *reminderService) GetReminder(ctx context.Context, id string) (models.Reminder, error) {
	var reminder models.Reminder
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return models.Reminder{}, err
	}
	query := s.db.Where("user_id = ?", userID).First(&reminder, id)
	if err := query.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Reminder{}, apperrors.ErrNotFound
		}
		return models.Reminder{}, apperrors.ErrInternal
	}
	return reminder, nil
}

func (s *reminderService) GetReminderByFilter(ctx context.Context, filter string) ([]models.Reminder, error) {
	var reminders []models.Reminder
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	query := s.db.Where("user_id = ?", userID).Where("message = ?", filter).Find(&reminders)
	if err := query.Error; err != nil {
		return nil, err
	}
	return reminders, nil
}

func (s *reminderService) ListDueReminders(ctx context.Context, now time.Time, limit int) ([]models.Reminder, error) {
	var reminders []models.Reminder

	query := s.db.WithContext(ctx).
		Where("time <= ? AND notified = ?", now, false).
		Order("time ASC").
		Limit(limit).
		Find(&reminders)
	if err := query.Error; err != nil {
		return nil, err
	}
	return reminders, nil
}

func (s *reminderService) MarkReminderNotified(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).
		Model(&models.Reminder{}).
		Where("id = ?", id).Update("notified", true)
	if result.Error != nil {
		return apperrors.ErrInternal
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
