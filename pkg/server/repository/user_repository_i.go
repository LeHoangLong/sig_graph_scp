package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type UserRepositoryI interface {
	// returns ErrNotFound if credentialChecker returns false
	FetchUserByUsernameAndCredentials(
		ctx context.Context,
		txId TransactionId,
		credentialChecker CredentialCheckerI,
		username string,
		credentials map[string]string,
	) (model_server.User, error)
	DoesUsernameExist(
		ctx context.Context,
		txId TransactionId,
		username string,
	) (bool, error)
	// returns ErrAlreadyExists if username already exists
	CreateUserWithUsernameAndCredentials(
		ctx context.Context,
		txId TransactionId,
		username string,
		credentials map[string]string,
	) (model_server.User, error)
}

type CredentialCheckerI interface {
	IsCredentialCorrect(ctx context.Context, username string, stored map[string]string, suppliedByUser map[string]string) (isCorrect bool, err error)
}
