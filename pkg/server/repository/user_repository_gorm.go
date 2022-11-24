package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
	"sig_graph_scp/pkg/utility"
)

type userRepositoryGorm struct {
	transactionManager *transactionManagerGorm
}

func NewUserRepositoryGorm(
	transactionManager *transactionManagerGorm,
) *userRepositoryGorm {
	return &userRepositoryGorm{
		transactionManager: transactionManager,
	}
}

type gormUser struct {
	ID uint64 `gorm:"primaryKey"`
}

type gormUsername struct {
	UserId   uint64 `gorm:"primaryKey"`
	Username string
}

type gormUserCredential struct {
	UserId          uint64 `gorm:"primaryKey"`
	CredentialName  string `gorm:"primaryKey"`
	CredentialValue string
}

func (r *userRepositoryGorm) FetchUserByUsernameAndCredentials(
	ctx context.Context,
	txId TransactionId,
	credentialChecker CredentialCheckerI,
	username string,
	credentials map[string]string,
) (model_server.User, error) {
	user, err := r.fetchUserByUsernameAndPassword(
		ctx,
		txId,
		credentialChecker,
		username,
		credentials,
	)

	if err != nil {
		return model_server.User{}, wrapError(err)
	}

	return user, nil
}

func (r *userRepositoryGorm) fetchUserByUsernameAndPassword(
	ctx context.Context,
	txId TransactionId,
	credentialChecker CredentialCheckerI,
	iUsername string,
	iCredentials map[string]string,
) (model_server.User, error) {
	tx, err := r.transactionManager.GetTransaction(ctx, txId)
	if err != nil {
		return model_server.User{}, err
	}

	username := gormUsername{}
	err = tx.Where("username = ?", iUsername).First(&username).Error
	if err != nil {
		return model_server.User{}, err
	}

	credentials := []gormUserCredential{}
	err = tx.Where("user_id = ?", username.UserId).Find(&credentials).Error
	if err != nil {
		return model_server.User{}, err
	}

	storedCredentials := map[string]string{}
	for i := range credentials {
		storedCredentials[credentials[i].CredentialName] = credentials[i].CredentialValue
	}

	isCorrect, err := credentialChecker.IsCredentialCorrect(ctx, iUsername, storedCredentials, iCredentials)
	if err != nil {
		return model_server.User{}, err
	}

	if !isCorrect {
		return model_server.User{}, utility.ErrNotFound
	}

	return model_server.User{
		ID: username.UserId,
	}, nil
}

func (r *userRepositoryGorm) DoesUsernameExist(
	ctx context.Context,
	txId TransactionId,
	username string,
) (bool, error) {
	exist, err := r.doesUsernameExist(
		ctx,
		txId,
		username,
	)

	if err != nil {
		return false, wrapError(err)
	}

	return exist, nil
}

func (r *userRepositoryGorm) doesUsernameExist(
	ctx context.Context,
	txId TransactionId,
	iUsername string,
) (bool, error) {
	tx, err := r.transactionManager.GetTransaction(ctx, txId)
	if err != nil {
		return false, nil
	}

	username := []gormUsername{}
	err = tx.Where("username = ?", iUsername).Find(&username).Error
	if err != nil {
		return false, err
	}

	if len(username) == 0 {
		return false, nil
	}

	return true, nil
}

func (r *userRepositoryGorm) CreateUserWithUsernameAndCredentials(
	ctx context.Context,
	txId TransactionId,
	username string,
	credentials map[string]string,
) (model_server.User, error) {
	user, err := r.createUserWithUsernameAndCredentials(
		ctx,
		txId,
		username,
		credentials,
	)

	if err != nil {
		return model_server.User{}, wrapError(err)
	}

	return user, nil
}

func (r *userRepositoryGorm) createUserWithUsernameAndCredentials(
	ctx context.Context,
	txId TransactionId,
	iUsername string,
	iCredentials map[string]string,
) (model_server.User, error) {
	tx, err := r.transactionManager.GetTransaction(ctx, txId)
	if err != nil {
		return model_server.User{}, err
	}

	user := gormUser{}
	err = tx.Create(&user).Error
	if err != nil {
		return model_server.User{}, err
	}

	username := gormUsername{
		UserId:   user.ID,
		Username: iUsername,
	}
	err = tx.Create(&username).Error
	if err != nil {
		return model_server.User{}, err
	}

	err = tx.Where("user_id = ?", user.ID).Delete(&gormUserCredential{}).Error
	if err != nil {
		return model_server.User{}, err
	}

	credentials := []gormUserCredential{}
	for credentialName, credentialValue := range iCredentials {
		credentials = append(credentials, gormUserCredential{
			UserId:          user.ID,
			CredentialName:  credentialName,
			CredentialValue: credentialValue,
		})
	}
	err = tx.Create(credentials).Error
	if err != nil {
		return model_server.User{}, err
	}

	return model_server.User{
		ID: user.ID,
	}, nil
}
