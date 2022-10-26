package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"
	"sig_graph_scp/pkg/utility"

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
	pagination PaginationOption[model_server.UserKeyPairId],
) ([]model_server.UserKeyPair, error) {
	keyPairs := []gormUserKeyPair{}
	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model_server.UserKeyPair{}, err
	}

	err = tx.Where("user_id = ? AND id >= ?", user.ID, pagination.MinId).Order("id asc").Limit(pagination.Limit).Find(&keyPairs).Error
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

func (r *userKeyRepositoryGorm) FetchUserWithPublicKey(
	ctx context.Context,
	transactionId TransactionId,
	publicKey string,
) (*model_server.User, error) {
	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return nil, err
	}

	gormKeyPair := gormUserKeyPair{}

	err = tx.Where("public_key = ?", publicKey).First(&gormKeyPair).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utility.ErrNotFound
		}

		return nil, err
	}

	return &model_server.User{
		ID: gormKeyPair.UserId,
	}, nil
}
