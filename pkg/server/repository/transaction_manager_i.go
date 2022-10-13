package repository_server

import "context"

type TransactionId uint64

type EIsolationLevel = int

const (
	EIsolationLevelReadUncommited EIsolationLevel = 0
	EIsolationLevelReadCommited   EIsolationLevel = 1
	EIsolationLevelRepeatableRead EIsolationLevel = 2
	EIsolationLevelSerializable   EIsolationLevel = 3
)

type TransactionOption struct {
	IsolationLevel EIsolationLevel
}

type TransactionManagerI interface {
	StartTransaction(ctx context.Context, option *TransactionOption) (TransactionId, error)
	Commit(ctx context.Context, id TransactionId) error
	Rollback(ctx context.Context, id TransactionId) error

	BypassTransaction(ctx context.Context) (TransactionId, error)
	StopBypassedTransaction(ctx context.Context, id TransactionId)
}
