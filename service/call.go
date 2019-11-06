package service

import (
	"context"

	"github.com/dmitrymomot/examplesrv/pb/examplesrv"
	"go.uber.org/zap"
)

func (s *service) Call(ctx context.Context, req *examplesrv.Req) (*examplesrv.Resp, error) {
	s.log.Infow("received string", zap.String("str", req.GetStr()))
	return &examplesrv.Resp{Str: "Received: " + req.GetStr()}, nil
}
