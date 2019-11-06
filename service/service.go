package service

import (
	"github.com/dmitrymomot/examplesrv/pb/examplesrv"
)

type service struct {
	// examplesrv.UnimplementedServiceServer
}

// New service factory
func New() examplesrv.ServiceServer {
	return &service{}
}
