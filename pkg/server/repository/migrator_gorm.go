package repository_server

import (
	"context"
	"fmt"
	"os"
	"path"

	"gorm.io/gorm"
)

type GormMigrationStepI interface {
	Up(ctx context.Context, transaction *gorm.DB) error
	Down(ctx context.Context, transaction *gorm.DB) error
}

type migratorGorm struct {
	versionRepository  *migratorVersionRepositoryGorm
	transactionManager *transactionManagerGorm
	steps              map[int]GormMigrationStepI
}

func NewMigratorGorm(
	versionRepository *migratorVersionRepositoryGorm,
	transactionManager *transactionManagerGorm,
) *migratorGorm {
	return &migratorGorm{
		versionRepository:  versionRepository,
		transactionManager: transactionManager,
		steps:              map[int]GormMigrationStepI{},
	}
}

func (m *migratorGorm) RegisterStep(stepNumber int, step GormMigrationStepI) error {
	m.steps[stepNumber] = step
	return nil
}

func (m *migratorGorm) Up(ctx context.Context, targetVersion uint32) error {
	for i := 0; i < int(targetVersion); i++ {
		err := func() error {
			txId, err := m.transactionManager.StartTransaction(ctx, &TransactionOption{
				IsolationLevel: EIsolationLevelSerializable,
			})
			defer m.transactionManager.Rollback(ctx, txId)

			version, err := m.versionRepository.GetVerionForUpdate(ctx, txId)
			if err != nil {
				return err
			}

			if version.Major > uint32(i) {
				return nil
			}

			tx, err := m.transactionManager.GetTransaction(ctx, txId)
			if err != nil {
				return err
			}

			file, err := os.ReadFile(path.Join(".", "pkg", "server", "migrations", fmt.Sprintf("migration_gorm_%d.up.sql", i+1)))
			if err != nil {
				return err
			}

			err = tx.Exec(string(file)).Error
			if err != nil {
				return err
			}

			version.Major += 1
			err = m.versionRepository.SetVerion(ctx, txId, version)
			if err != nil {
				return err
			}

			m.transactionManager.Commit(ctx, txId)
			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

func (m *migratorGorm) Down(ctx context.Context, targetVersion uint32) error {
	currentVersion := 0
	err := func() error {
		txId, err := m.transactionManager.StartTransaction(ctx, &TransactionOption{
			IsolationLevel: EIsolationLevelSerializable,
		})
		if err != nil {
			return err
		}
		defer m.transactionManager.Rollback(ctx, txId)
		version, err := m.versionRepository.GetVerionForUpdate(ctx, txId)
		currentVersion = int(version.Major)
		return nil
	}()
	if err != nil {
		return err
	}

	for i := currentVersion; i > int(targetVersion) && i > 0; i-- {
		err := func() error {
			txId, err := m.transactionManager.StartTransaction(ctx, &TransactionOption{
				IsolationLevel: EIsolationLevelSerializable,
			})
			defer m.transactionManager.Rollback(ctx, txId)

			version, err := m.versionRepository.GetVerionForUpdate(ctx, txId)
			if err != nil {
				return err
			}

			if version.Major > uint32(i) {
				return nil
			}

			tx, err := m.transactionManager.GetTransaction(ctx, txId)
			if err != nil {
				return err
			}

			file, err := os.ReadFile(path.Join(".", "pkg", "server", "migrations", fmt.Sprintf("migration_gorm_%d.down.sql", i)))
			if err != nil {
				return err
			}

			err = tx.Exec(string(file)).Error
			if err != nil {
				return err
			}

			version.Major -= 1
			err = m.versionRepository.SetVerion(ctx, txId, version)
			if err != nil {
				return err
			}

			m.transactionManager.Commit(ctx, txId)
			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}
