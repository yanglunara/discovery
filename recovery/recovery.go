package recovery

import (
	"context"
	"errors"
	"runtime"
	"time"

	"github.com/yanglunara/discovery/transport/middleware"
	log "github.com/yunbaifan/pkg/logger"
	"go.uber.org/zap"
)

var (
	ErrUnknownRequest = errors.New("unknown request")
)

type Latency struct {
}

type HandlerFunc func(ctx context.Context, req, err interface{}) error

type Option func(*option)

type option struct {
	handler HandlerFunc
}

func Recovery(opts ...Option) middleware.Middleware {
	op := option{
		handler: func(ctx context.Context, req, err interface{}) error {
			return ErrUnknownRequest
		},
	}
	for _, o := range opts {
		o(&op)
	}
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (resp interface{}, err error) {
			startTime := time.Now()
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 64<<10)
					buf = buf[:runtime.Stack(buf, false)]
					log.FromZapLoggerContext(ctx).Error("recover err ",
						zap.Error(err),
						zap.Any("req", req),
						zap.ByteString("stack", buf),
					)
					ctx = context.WithValue(ctx, Latency{}, time.Since(startTime).Seconds())
					err = op.handler(ctx, req, r)
				}
			}()
			return next(ctx, req)
		}
	}
}
