package controller_server

import (
	"context"
	"fmt"
	model_server "sig_graph_scp/pkg/server/model"
	repository_server "sig_graph_scp/pkg/server/repository"

	"golang.org/x/crypto/bcrypt"
)

type userController struct {
	userRepository     repository_server.UserRepositoryI
	transactionManager repository_server.TransactionManagerI
}

func NewUserController(
	userRepository repository_server.UserRepositoryI,
	transactionManager repository_server.TransactionManagerI,
) *userController {
	return &userController{
		userRepository:     userRepository,
		transactionManager: transactionManager,
	}
}

func (c *userController) CreateUserWithUsernameAndPassword(
	ctx context.Context,
	username string,
	password string,
) (model_server.User, error) {
	txId, err := c.transactionManager.StartTransaction(
		ctx,
		&repository_server.TransactionOption{
			IsolationLevel: repository_server.EIsolationLevelReadCommited,
		},
	)
	if err != nil {
		return model_server.User{}, err
	}
	defer c.transactionManager.Rollback(ctx, txId)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user, err := c.userRepository.CreateUserWithUsernameAndCredentials(
		ctx,
		txId,
		username,
		map[string]string{
			"type": "bcrypt",
			"hash": string(hashedPassword),
		},
	)
	if err != nil {
		fmt.Println("err: ", err)
		return model_server.User{}, err
	}

	err = c.transactionManager.Commit(ctx, txId)
	if err != nil {
		return model_server.User{}, err
	}

	return user, nil
}

func (c *userController) FetchUserWithusernameAndPassword(
	ctx context.Context,
	username string,
	password string,
) (model_server.User, error) {
	txId, err := c.transactionManager.BypassTransaction(
		ctx,
	)
	if err != nil {
		return model_server.User{}, err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, txId)

	user, err := c.userRepository.FetchUserByUsernameAndCredentials(
		ctx,
		txId,
		c,
		username,
		map[string]string{
			"password": password,
		},
	)
	if err != nil {
		return model_server.User{}, err
	}

	return user, nil
}

func (c *userController) IsCredentialCorrect(
	ctx context.Context,
	username string,
	storedCredential map[string]string,
	userSuppliedCredential map[string]string,
) (bool, error) {
	if storedCredential["type"] != "bcrypt" {
		return false, nil
	}

	hash := storedCredential["hash"]
	password := userSuppliedCredential["password"]
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false, nil
	} else {
		return true, nil
	}
}
