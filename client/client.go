package client

import (
	"fmt"

	"github.com/dmitrymomot/examplesrv/pb/examplesrv"
	"google.golang.org/grpc"
)

type (
	// Client interface
	Client interface {
		examplesrv.ServiceClient
		Close() error
	}

	client struct {
		examplesrv.ServiceClient
		conn *grpc.ClientConn
	}
)

// Close gRPC connection with examplesrv service
func (c *client) Close() error {
	return c.conn.Close()
}

// NewDefaultClient returns grpc connection
func NewDefaultClient(addr string) (Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("did not connect to %s: %v", addr, err)
	}
	c := examplesrv.NewServiceClient(conn)
	return &client{c, conn}, nil
}
