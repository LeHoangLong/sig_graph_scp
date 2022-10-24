package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormVersionModel struct {
	ID    uint32 `gorm:"primaryKey;check:id=1"`
	Major uint32
}

type migratorVersionRepositoryGorm struct {
	transactionManager transactionManagerGorm
}

func NewMigratorVersionRepositoryGorm(
	transactionManager transactionManagerGorm,
) migratorVersionRepositoryGorm {
	return migratorVersionRepositoryGorm{
		transactionManager: transactionManager,
	}
}

func (r *migratorVersionRepositoryGorm) CreatTableIfNotExists(ctx context.Context) error {
	txId, err := r.transactionManager.StartTransaction(ctx, nil)
	if err != nil {
		return err
	}
	defer r.transactionManager.Rollback(ctx, txId)

	transaction, err := r.transactionManager.GetTransaction(ctx, txId)
	if err != nil {
		return err
	}

	err = transaction.Exec(`
		CREATE TABLE IF NOT EXISTS "gorm_version_models" (
			id SERIAL PRIMARY KEY CHECK (id = 1),
			major INTEGER NOT NULL
		) 
	`).Error

	if err != nil {
		return err
	}

	r.transactionManager.Commit(ctx, txId)
	return nil
}

func (r *migratorVersionRepositoryGorm) GetVerionForUpdate(ctx context.Context, transactionId TransactionId) (model_server.MigrationVerion, error) {
	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return model_server.MigrationVerion{}, err
	}

	version := gormVersionModel{
		ID: 1,
	}

	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&version).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model_server.MigrationVerion{
				Major: 0,
			}, nil
		} else {
			return model_server.MigrationVerion{}, err
		}
	}

	modelVersion := model_server.MigrationVerion{
		Major: version.Major,
	}
	return modelVersion, nil
}

func (r *migratorVersionRepositoryGorm) SetVerion(ctx context.Context, transactionId TransactionId, version model_server.MigrationVerion) error {
	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return err
	}

	gormVersion := gormVersionModel{
		ID:    1,
		Major: version.Major,
	}

	err = tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&gormVersion).Error
	if err != nil {
		return err
	}

	return nil
}
