package endpoint

import (
	"context"
	apperrors "reminder/errors"
	"reminder/models"
	"reminder/service"
	"strconv"
	"time"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/log"
)

type SetReminderRequest struct {
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

type SetReminderResponse struct {
	Reminder models.Reminder `json:"reminder"`
}

type GetRemindersRequest struct {
	Limit   int  `json:"limit"`
	AfterID uint `json:"after_id"`
}

type GetRemindersResponse struct {
	Reminders  []models.Reminder `json:"reminders"`
	Limit      int               `json:"limit"`
	HasMore    bool              `json:"has_more"`
	NextCursor string            `json:"next_cursor,omitempty"`
}

type GetReminderRequest struct {
	ID string `json:"id"`
}

type GetReminderResponse struct {
	Reminder models.Reminder `json:"reminder"`
}

type GetReminderByFilterRequest struct {
	Filter string `json:"filter"`
}

type GetReminderByFilterResponse struct {
	Reminders []models.Reminder `json:"reminders"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	User models.User `json:"user"`
}

func MakeRegisterEndpoint(authSvc service.AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(RegisterRequest)
		if req.Username == "" || req.Password == "" {
			return nil, apperrors.ErrInvalidInput
		}
		user, err := authSvc.Register(ctx, req.Username, req.Password)
		if err != nil {
			return nil, err
		}
		return RegisterResponse{User: user}, nil
	}
}

func MakeLoginEndpoint(authSvc service.AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(LoginRequest)
		if req.Username == "" || req.Password == "" {
			return nil, apperrors.ErrInvalidInput
		}
		token, err := authSvc.Login(ctx, req.Username, req.Password)
		if err != nil {
			return nil, err
		}
		return LoginResponse{Token: token}, nil
	}
}

func MakeSetReminderEndpoint(service service.ReminderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SetReminderRequest)
		reminder := models.Reminder{
			Message: req.Message,
			Time:    req.Time,
		}
		created, err := service.SetReminder(ctx, reminder)
		if err != nil {
			return nil, err
		}
		return SetReminderResponse{Reminder: created}, nil
	}
}

func MakeGetRemindersEndpoint(service service.ReminderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetRemindersRequest)
		result, err := service.GetReminders(ctx, req.AfterID, req.Limit)
		if err != nil {
			return nil, err
		}
		nextCursor := ""
		if result.HasMore && len(result.Reminders) > 0 {
			last := result.Reminders[len(result.Reminders)-1]
			nextCursor = strconv.FormatUint(uint64(last.ID), 10)
		}

		return GetRemindersResponse{Reminders: result.Reminders, Limit: req.Limit, HasMore: result.HasMore, NextCursor: nextCursor}, nil
	}
}

func MakeGetReminderEndpoint(service service.ReminderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetReminderRequest)
		reminder, err := service.GetReminder(ctx, req.ID)
		if err != nil {
			return nil, err
		}
		return GetReminderResponse{Reminder: reminder}, nil
	}
}

func MakeGetReminderByFilterEndpoint(service service.ReminderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetReminderByFilterRequest)
		reminders, err := service.GetReminderByFilter(ctx, req.Filter)
		if err != nil {
			return nil, err
		}
		return GetReminderByFilterResponse{Reminders: reminders}, nil
	}
}

type Endpoints struct {
	SetReminderEndpoint         endpoint.Endpoint
	GetRemindersEndpoint        endpoint.Endpoint
	GetReminderEndpoint         endpoint.Endpoint
	GetReminderByFilterEndpoint endpoint.Endpoint
	LoginEndpoint               endpoint.Endpoint
	RegisterEndpoint            endpoint.Endpoint
}

func MakeEndpoints(service service.ReminderService, logger kitlog.Logger, tokens TokenParser, authSvc service.AuthService) Endpoints {
	return Endpoints{
		SetReminderEndpoint:         chainAuthLog(tokens, logger, "SetReminder", MakeSetReminderEndpoint(service)),
		GetRemindersEndpoint:        chainAuthLog(tokens, logger, "GetReminders", MakeGetRemindersEndpoint(service)),
		GetReminderEndpoint:         chainAuthLog(tokens, logger, "GetReminder", MakeGetReminderEndpoint(service)),
		GetReminderByFilterEndpoint: chainAuthLog(tokens, logger, "GetReminderByFilter", MakeGetReminderByFilterEndpoint(service)),
		LoginEndpoint:               LoggingMiddleware(logger, "login")(MakeLoginEndpoint(authSvc)),
		RegisterEndpoint:            LoggingMiddleware(logger, "register")(MakeRegisterEndpoint(authSvc)),
	}
}

func chainAuthLog(parser TokenParser, logger kitlog.Logger, name string, ep endpoint.Endpoint) endpoint.Endpoint {
	return AuthMiddleware(parser)(LoggingMiddleware(logger, name)(ep))
}
