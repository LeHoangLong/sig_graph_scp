package utility

import (
	"context"
)

type MutexI interface {
	Lock(ctx context.Context) bool
	Unlock(ctx context.Context)
}

type mutex struct {
	lock chan bool
}

func NewMutex() *mutex {
	return &mutex{
		lock: make(chan bool, 1),
	}
}

func (m *mutex) Lock(ctx context.Context) bool {
	select {
	case m.lock <- true:
		return true
	case <-ctx.Done():
		return false
	}
}
func (m *mutex) Unlock(ctx context.Context) {
	if len(m.lock) > 0 {
		<-m.lock
	}
}
