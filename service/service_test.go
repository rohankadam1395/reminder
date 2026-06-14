package service

import (
	"context"
	"errors"
	apperrors "reminder/errors"
	"reminder/models"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Reminder{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestSetReminderAndGetReminder(t *testing.T) {
	svc := NewReminderService(newTestDB(t))
	ctx := context.Background()

	created, err := svc.SetReminder(ctx, models.Reminder{Message: "hi", Time: time.Now().Add(1 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 {
		t.Fatal("expected ID after create")
	}

	got, err := svc.GetReminder(ctx, "1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Message != "hi" {
		t.Fatalf("got %+v", got)
	}
}

func TestGetReminder_NotFound(t *testing.T) {
	svc := NewReminderService(newTestDB(t))
	_, err := svc.GetReminder(context.Background(), "999")
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestGetReminders_HasMore(t *testing.T) {
	db := newTestDB(t)
	svc := NewReminderService(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if _, err := svc.SetReminder(ctx, models.Reminder{Message: "m", Time: time.Now().Add(time.Duration(i) * time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}

	page1, err := svc.GetReminders(ctx, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !page1.HasMore || len(page1.Reminders) != 2 {
		t.Fatalf("page1: %v", page1)
	}

	lastID := page1.Reminders[1].ID
	page2, err := svc.GetReminders(ctx, lastID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if page2.HasMore || len(page2.Reminders) != 1 {
		t.Fatalf("page2: %+v", page2)
	}
}

func TestListDueReminders(t *testing.T) {
	svc := NewReminderService(newTestDB(t))
	ctx := context.Background()
	now := time.Now()

	past, _ := svc.SetReminder(ctx, models.Reminder{
		Message: "due",
		Time:    now.Add(-time.Minute),
	})
	svc.SetReminder(ctx, models.Reminder{
		Message: "future",
		Time:    now.Add(time.Hour),
	})

	due, err := svc.ListDueReminders(ctx, now, 10)
	if err != nil || len(due) != 1 || due[0].ID != past.ID {
		t.Fatalf("got %+v, err=%v", due, err)
	}

}
