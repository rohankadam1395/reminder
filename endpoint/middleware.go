package endpoint

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/log"
)

func LoggingMiddleware(logger kitlog.Logger, name string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			defer func(begin time.Time) {
				_ = logger.Log(
					"endpoint", name,
					"took", time.Since(begin),
				)
			}(time.Now())

			response, err := next(ctx, request)
			if err != nil {
				_ = logger.Log(
					"endpoint", name, "err", err,
				)
			}
			return response, err
		}
	}

}
