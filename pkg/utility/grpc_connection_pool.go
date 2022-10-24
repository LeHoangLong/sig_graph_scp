package utility

import (
	"context"

	"google.golang.org/grpc"
)

type grpcConnectionPool struct {
	connections map[string][]*grpc.ClientConn
	mtx         MutexI
}

func NewGrpcConnectionPool() *grpcConnectionPool {
	return &grpcConnectionPool{
		connections: map[string][]*grpc.ClientConn{},
		mtx:         NewMutex(),
	}
}

func (p *grpcConnectionPool) NewConnection(
	ctx context.Context,
	url string,
) (*grpc.ClientConn, error) {
	if !p.mtx.Lock(ctx) {
		return nil, ErrTimedOut
	}

	defer p.mtx.Unlock(ctx)
	if availableConnections, ok := p.connections[url]; ok {
		if len(availableConnections) > 0 {
			connection := availableConnections[len(availableConnections)-1]
			availableConnections = availableConnections[0 : len(availableConnections)-1]
			return connection, nil
		}
	} else {
		p.connections[url] = []*grpc.ClientConn{}
	}

	clientConn, err := grpc.DialContext(ctx, url, grpc.WithBlock)
	if err != nil {
		return nil, err
	}

	return clientConn, nil
}

func (p *grpcConnectionPool) ReturnConnection(ctx context.Context, url string, conn *grpc.ClientConn) error {
	if !p.mtx.Lock(ctx) {
		return ErrTimedOut
	}

	defer p.mtx.Unlock(ctx)

	if _, ok := p.connections[url]; !ok {
		p.connections[url] = []*grpc.ClientConn{}
	}

	p.connections[url] = append(p.connections[url], conn)
	return nil
}
