package repository_server

import (
	"context"
	"sig_graph_scp/pkg/model"
)

type MigratorVersionRepositoryI interface {
	GetVerionForUpdate(ctx context.Context, transactionId TransactionId) (model.MigrationVerion, error)
	SetVerion(ctx context.Context, transactionId TransactionId, version model.MigrationVerion) error
}
