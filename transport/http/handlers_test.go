package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reminder/endpoint"
	apperrors "reminder/errors"
	"testing"

	"github.com/gorilla/mux"
)

func TestEncodeError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{"not found", apperrors.ErrNotFound, http.StatusNotFound, "reminder not found"},
		{"invalid input", apperrors.ErrInvalidInput, http.StatusBadRequest, "invalid input"},
		{"internal", apperrors.ErrInternal, http.StatusInternalServerError, "internal server error"},
		{"unknown", errors.New("unknown error"), http.StatusInternalServerError, "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			encodeError(context.Background(), tt.err, rec)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status got %d, want %d", rec.Code, tt.wantStatus)
			}

			var body map[string]string
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode body: %v", err)
			}

			if body["error"] != tt.wantMsg {
				t.Fatalf("message got %s want %s", body["error"], tt.wantMsg)
			}
		})
	}
}

func TestDecodeGetRemindersRequest(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/get-reminders", nil)
		got, err := decodeGetRemindersRequest(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		r := got.(endpoint.GetRemindersRequest)
		if r.Limit != 10 || r.AfterID != 0 {
			t.Fatalf("got %v", r)
		}
	})

	t.Run("invalid limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/get-reminders?limit=abc", nil)
		_, err := decodeGetRemindersRequest(context.Background(), req)
		if !errors.Is(err, apperrors.ErrInvalidInput) {
			t.Fatalf("got %v", err)
		}
	})
}

func TestDecodeGetReminderRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/get-reminders/5", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "5"})

	got, err := decodeGetReminderRequest(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if got.(endpoint.GetReminderRequest).ID != "5" {
		t.Fatalf("got %+v", got)
	}
}
