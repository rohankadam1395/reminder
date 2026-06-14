package test

import (
	"net/http"
	"net/http/httptest"
	"reminder/endpoint"
	"reminder/models"
	"reminder/service"
	httptransport "reminder/transport/http"
	"testing"

	kitlog "github.com/go-kit/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHttp_GetReminder_NotFound(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Reminder{})

	svc := service.NewReminderService(db)
	handler := httptransport.NewHTTPHandler(endpoint.MakeEndpoints(svc, kitlog.NewNopLogger(), nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/get-reminder/1", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: %d body: %s", rec.Code, rec.Body.String())
	}

}
