package main

import (
	"log"
	"net"
	"os"

	"github.com/dmitrymomot/examplesrv/pb/examplesrv"
	"github.com/dmitrymomot/examplesrv/service"
	"google.golang.org/grpc"
)

var buildTag = "dev"

func main() {
	lis, err := net.Listen("tcp", os.Getenv("GRPC_PORT"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	srv := service.New()
	examplesrv.RegisterServiceServer(s, srv)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
