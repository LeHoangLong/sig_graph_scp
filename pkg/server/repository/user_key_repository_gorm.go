package repository_server

import (
	"context"
	"sig_graph_scp/pkg/model"
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
	ID         uint64       `gorm:"primaryKey"`
	UserId     model.UserId `gorm:"index:user_id_index"`
	User       model.User   `gorm:"foreignKey:UserId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PublicKey  string
	PrivateKey string
}

func (r *userKeyRepositoryGorm) FetchKeyPairsOfUser(
	ctx context.Context,
	transactionId TransactionId,
	user *model.User,
) ([]model.UserKeyPair, error) {
	keyPairs := []gormUserKeyPair{}
	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model.UserKeyPair{}, err
	}

	err = tx.Where("user_id = ?", user.ID).Find(&keyPairs).Error
	if err != nil {
		return []model.UserKeyPair{}, err
	}

	modelKeyPairs := []model.UserKeyPair{}
	for i := range keyPairs {
		modelKeyPair := model.UserKeyPair{
			Id:      keyPairs[i].ID,
			Public:  keyPairs[i].PublicKey,
			Private: keyPairs[i].PrivateKey,
		}

		modelKeyPairs = append(modelKeyPairs, modelKeyPair)
	}

	return modelKeyPairs, nil
}

func (r *userKeyRepositoryGorm) AddKeyPairToUser(ctx context.Context, transactionId TransactionId, user *model.User, keyPair *model.UserKeyPair) error {
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
