package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"reminder/endpoint"
	apperrors "reminder/errors"
	"strconv"
	"strings"

	"reminder/auth"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func authRequestContext(ctx context.Context, r *http.Request) context.Context {
	header := r.Header.Get("Authorization")

	if strings.HasPrefix(header, "Bearer ") {
		token := strings.TrimPrefix(header, "Bearer ")
		ctx = auth.WithBearerToken(ctx, token)
	}
	return ctx
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	var status int
	var message string

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		status = http.StatusNotFound
		message = "reminder not found"
	case errors.Is(err, apperrors.ErrInvalidInput):
		status = http.StatusBadRequest
		message = "invalid input"
	case errors.Is(err, apperrors.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = "unauthorized"
	case errors.Is(err, apperrors.ErrForbidden):
		status = http.StatusForbidden
	default:
		status = http.StatusInternalServerError
		message = "internal server error"
		log.Printf("internal server error: %v", err)
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func serverOptions() []httptransport.ServerOption {
	return []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
		httptransport.ServerBefore(authRequestContext),
	}
}

func decodeSetReminderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request endpoint.SetReminderRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, apperrors.ErrInvalidInput
	}
	if request.Message == "" || request.Time.IsZero() {
		return nil, apperrors.ErrInvalidInput
	}
	return request, nil
}

func decodeGetRemindersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	q := r.URL.Query()

	limit := 10
	if l := q.Get("limit"); l != "" {
		v, err := strconv.Atoi(l)
		if err != nil || v < 1 || v > 100 {
			return nil, apperrors.ErrInvalidInput
		}
		limit = v
	}

	var afterID uint
	if c := q.Get("after_id"); c != "" {
		v, err := strconv.ParseUint(c, 10, 64)
		if err != nil {
			return nil, apperrors.ErrInvalidInput
		}
		afterID = uint(v)
	}

	return endpoint.GetRemindersRequest{
		Limit:   limit,
		AfterID: afterID,
	}, nil
}

func decodeGetReminderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id := mux.Vars(r)["id"]
	if id == "" {
		return nil, apperrors.ErrInvalidInput
	}
	return endpoint.GetReminderRequest{ID: id}, nil
}

func decodeGetReminderByFilterRequest(_ context.Context, r *http.Request) (interface{}, error) {
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		return nil, apperrors.ErrInvalidInput
	}
	return endpoint.GetReminderByFilterRequest{Filter: filter}, nil
}

func decodeLoginRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, apperrors.ErrInvalidInput
	}
	return req, nil
}

func decodeRegisterRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, apperrors.ErrInvalidInput
	}
	if req.Username == "" || req.Password == "" {
		return nil, apperrors.ErrInvalidInput
	}
	return req, nil
}

func NewHTTPHandler(endpoints endpoint.Endpoints) http.Handler {
	m := mux.NewRouter()
	opts := serverOptions()

	m.Handle("/set-reminder", httptransport.NewServer(
		endpoints.SetReminderEndpoint,
		decodeSetReminderRequest,
		encodeResponse,
		opts...,
	)).Methods(http.MethodPost)

	m.Handle("/get-reminder/{id}", httptransport.NewServer(
		endpoints.GetReminderEndpoint,
		decodeGetReminderRequest,
		encodeResponse,
		opts...,
	)).Methods(http.MethodGet)

	m.Handle("/get-reminders", httptransport.NewServer(
		endpoints.GetRemindersEndpoint,
		decodeGetRemindersRequest,
		encodeResponse,
		opts...,
	)).Methods(http.MethodGet)

	m.Handle("/get-reminder-by-filter", httptransport.NewServer(
		endpoints.GetReminderByFilterEndpoint,
		decodeGetReminderByFilterRequest,
		encodeResponse,
		opts...,
	)).Methods(http.MethodGet)

	m.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	m.Handle("/login", httptransport.NewServer(
		endpoints.LoginEndpoint,
		decodeLoginRequest,
		encodeResponse,
		opts...,
	)).Methods(http.MethodPost)

	m.Handle("/register", httptransport.NewServer(
		endpoints.RegisterEndpoint,
		decodeRegisterRequest,
		encodeResponse,
		opts...,
	)).Methods(http.MethodPost)

	return m
}
