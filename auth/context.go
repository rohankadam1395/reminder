package auth

import "context"

type contextKey string

const userIDKey contextKey = "userID"
const bearerTokenKey contextKey = "bearerToken"

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok && v != ""
}

func WithBearerToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, bearerTokenKey, token)
}

func BearerTokenFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(bearerTokenKey).(string)
	return v, ok && v != ""
}
