package utility

import (
	"context"

	"google.golang.org/grpc"
)

type GrpcConnectionPoolI interface {
	NewConnection(ctx context.Context, url string) (*grpc.ClientConn, error)
	ReturnConnection(ctx context.Context, url string, conn *grpc.ClientConn) error
}
