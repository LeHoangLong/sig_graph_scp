package controller_server

import (
	"context"
	"sig_graph_scp/pkg/model"
	repository_server "sig_graph_scp/pkg/server/repository"
)

type userKeyPairController struct {
	repository         repository_server.UserKeyRepositoryI
	transactionManager repository_server.TransactionManagerI
}

func NewUserKeyPairController(
	repository repository_server.UserKeyRepositoryI,
	transactionManager repository_server.TransactionManagerI,
) UserKeyPairControllerI {
	return &userKeyPairController{
		repository:         repository,
		transactionManager: transactionManager,
	}
}

func (c *userKeyPairController) FetchKeyPairsByUser(ctx context.Context, user *model.User) ([]model.UserKeyPair, error) {
	tx, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, tx)

	keyPairs, err := c.repository.FetchKeyPairsOfUser(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	return keyPairs, nil
}

func (c *userKeyPairController) AddKeyPairToUser(ctx context.Context, user *model.User, keyPair *model.UserKeyPair) error {
	tx, err := c.transactionManager.BypassTransaction(ctx)
	if err != nil {
		return err
	}
	defer c.transactionManager.StopBypassedTransaction(ctx, tx)

	err = c.repository.AddKeyPairToUser(ctx, tx, user, keyPair)
	return err
}
