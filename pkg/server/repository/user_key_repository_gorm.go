package repository_server

import (
	"context"
	"sig_graph_scp/pkg/model"

	"gorm.io/gorm"
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
	gorm.Model
	UserId     model.UserId `gorm:"primaryKey,priority:1"`
	ID         uint64       `gorm:"primaryKey,priority:2"`
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
			Public:  keyPairs[i].PublicKey,
			Private: keyPairs[i].PrivateKey,
		}

		modelKeyPairs = append(modelKeyPairs, modelKeyPair)
	}

	return modelKeyPairs, nil
}
