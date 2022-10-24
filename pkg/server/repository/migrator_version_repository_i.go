package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type MigratorVersionRepositoryI interface {
	GetVerionForUpdate(ctx context.Context, transactionId TransactionId) (model_server.MigrationVerion, error)
	SetVerion(ctx context.Context, transactionId TransactionId, version model_server.MigrationVerion) error
}
