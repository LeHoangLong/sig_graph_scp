package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
)

type userKeyRepositoryGorm struct {
	transactionManager *transactionManagerGorm
}

func NewUserKeyRepositoryGorm(
	transactionManager *transactionManagerGorm,
) *userKeyRepositoryGorm {
	return &userKeyRepositoryGorm{
		transactionManager: transactionManager,
	}
}

type gormUserKeyPair struct {
	ID         uint64              `gorm:"primaryKey"`
	UserId     model_server.UserId `gorm:"index:user_id_index"`
	User       model_server.User   `gorm:"foreignKey:UserId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PublicKey  string
	PrivateKey string
}

func (r *userKeyRepositoryGorm) FetchKeyPairsOfUser(
	ctx context.Context,
	transactionId TransactionId,
	user *model_server.User,
) ([]model_server.UserKeyPair, error) {
	keyPairs := []gormUserKeyPair{}
	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model_server.UserKeyPair{}, err
	}

	err = tx.Where("user_id = ?", user.ID).Find(&keyPairs).Error
	if err != nil {
		return []model_server.UserKeyPair{}, err
	}

	modelKeyPairs := []model_server.UserKeyPair{}
	for i := range keyPairs {
		modelKeyPair := model_server.UserKeyPair{
			Id:      keyPairs[i].ID,
			Public:  keyPairs[i].PublicKey,
			Private: keyPairs[i].PrivateKey,
		}

		modelKeyPairs = append(modelKeyPairs, modelKeyPair)
	}

	return modelKeyPairs, nil
}

func (r *userKeyRepositoryGorm) AddKeyPairToUser(ctx context.Context, transactionId TransactionId, user *model_server.User, keyPair *model_server.UserKeyPair) error {
	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return err
	}

	gormKeyPair := gormUserKeyPair{
		UserId:     user.ID,
		PublicKey:  keyPair.Public,
		PrivateKey: keyPair.Private,
	}

	err = tx.Omit("User").Create(&gormKeyPair).Error
	return err
}
