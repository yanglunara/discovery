package middleware

import "context"

type (
	Handler func(ctx context.Context, req interface{}) (interface{}, error)

	Middleware func(Handler) Handler
)

func Next(m ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](next)
		}
		return next
	}
}
