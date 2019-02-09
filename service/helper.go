package service

import (
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func logging(ctx context.Context) *zap.SugaredLogger {
	return ctx.Value("log").(*zap.SugaredLogger)
}
