package endpoint

import (
	"context"
	"reminder/auth"
	apperrors "reminder/errors"

	"github.com/go-kit/kit/endpoint"
)

type TokenParser interface {
	Parse(tokenString string) (string, error)
}

func AuthMiddleware(parser TokenParser) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			token, ok := auth.BearerTokenFromContext(ctx)
			if !ok {
				return nil, apperrors.ErrUnauthorized
			}
			userID, err := parser.Parse(token)
			if err != nil {
				return nil, apperrors.ErrUnauthorized
			}
			ctx = auth.WithUserID(ctx, userID)
			return next(ctx, request)
		}
	}
}
