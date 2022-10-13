package repository_server

import (
	"context"
	"database/sql"
	"fmt"
	"sig_graph_scp/pkg/utility"

	"gorm.io/gorm"
)

type transactionManagerGorm struct {
	db                 *gorm.DB
	mtx                utility.MutexI
	transactionCounter uint64
	activeTransactions map[TransactionId]*gorm.DB
}

func NewTransactionManagerGorm(db *gorm.DB) *transactionManagerGorm {
	return &transactionManagerGorm{
		db: db,
	}
}

func isolationLevelToGormIsolationLevel(level EIsolationLevel) sql.IsolationLevel {
	switch level {
	case EIsolationLevelReadUncommited:
		return sql.LevelReadUncommitted
	case EIsolationLevelReadCommited:
		return sql.LevelReadCommitted
	case EIsolationLevelRepeatableRead:
		return sql.LevelRepeatableRead
	case EIsolationLevelSerializable:
		return sql.LevelSerializable
	}

	return sql.LevelDefault
}

func (m *transactionManagerGorm) StartTransaction(ctx context.Context, option *TransactionOption) (TransactionId, error) {
	if !m.mtx.Lock(ctx) {
		return 0, fmt.Errorf("%w: could not get mutex", utility.ErrTimedOut)
	}
	defer m.mtx.Unlock(ctx)

	transactionId := TransactionId(m.transactionCounter + 1)
	m.transactionCounter++
	if m.transactionCounter == 0 {
		m.transactionCounter = 1
	}
	tx := m.db.Begin()
	m.activeTransactions[transactionId] = tx

	return transactionId, nil
}

func (m *transactionManagerGorm) Commit(ctx context.Context, id TransactionId) error {
	if !m.mtx.Lock(ctx) {
		return fmt.Errorf("%w: could not get mutex", utility.ErrTimedOut)
	}
	defer m.mtx.Unlock(ctx)

	if transaction, ok := m.activeTransactions[id]; ok {
		err := transaction.Commit().Error
		delete(m.activeTransactions, id)
		return err
	}

	return nil
}

func (m *transactionManagerGorm) Rollback(ctx context.Context, id TransactionId) error {
	if !m.mtx.Lock(ctx) {
		return fmt.Errorf("%w: could not get mutex", utility.ErrTimedOut)
	}
	defer m.mtx.Unlock(ctx)

	if transaction, ok := m.activeTransactions[id]; ok {
		err := transaction.Rollback().Error
		delete(m.activeTransactions, id)
		return err
	}

	return nil

}

func (m *transactionManagerGorm) BypassTransaction(ctx context.Context) (TransactionId, error) {
	return 0, nil
}

func (m *transactionManagerGorm) StopBypassedTransaction(ctx context.Context, id TransactionId) {

}

func (m *transactionManagerGorm) GetTransaction(ctx context.Context, id TransactionId) (*gorm.DB, error) {
	if id == 0 {
		return m.db, nil
	}

	if transaction, ok := m.activeTransactions[id]; !ok {
		return nil, utility.ErrNotFound
	} else {
		return transaction, nil
	}
}
