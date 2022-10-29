package main

import (
	"context"
	"fmt"
	"os"
	repository_server "sig_graph_scp/pkg/server/repository"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// init db
	connectionStr := os.Getenv("DB_CONNECTION")
	if connectionStr == "" {
		panic("missing DB_CONNECTION env var")
	}

	db, err := gorm.Open(postgres.Open(connectionStr), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("could not connect to db: %s", err))
	}

	// transaction manager
	transactionManager := repository_server.NewTransactionManagerGorm(db)

	// migrate database
	versionRepository := repository_server.NewMigratorVersionRepositoryGorm(*transactionManager)
	{
		ctx := context.Background()
		err := versionRepository.CreatTableIfNotExists(ctx)
		if err != nil {
			panic(fmt.Sprintf("could not create verion table: %s", err))
		}
	}
	migrator := repository_server.NewMigratorGorm(&versionRepository, transactionManager)
	{
		ctx := context.Background()
		err := migrator.Down(ctx, 2)
		if err != nil {
			panic(fmt.Sprintf("could not migrate database: %s", err))
		}
	}
}
