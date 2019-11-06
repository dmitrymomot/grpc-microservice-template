package service

import (
	"context"
	"log"

	"github.com/dmitrymomot/examplesrv/pb/examplesrv"
)

func (*service) Call(ctx context.Context, req *examplesrv.Req) (*examplesrv.Resp, error) {
	log.Println("Received: " + req.GetStr())
	return &examplesrv.Resp{Str: "Received: " + req.GetStr()}, nil
}
