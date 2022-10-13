package repository_server

import (
	"context"
	"sig_graph_scp/pkg/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type assetRepositoryGorm struct {
	transactionManagerGorm *transactionManagerGorm
	nodeRepository         NodeRepositoryI
}

var _ AssetRepositoryI = (*assetRepositoryGorm)(nil)

func NewAssetRepositoryGorm(
	transactionManagerGorm *transactionManagerGorm,
	nodeRepository NodeRepositoryI,
) *assetRepositoryGorm {
	return &assetRepositoryGorm{
		transactionManagerGorm: transactionManagerGorm,
		nodeRepository:         nodeRepository,
	}
}

type gormAsset struct {
	gorm.Model

	CreationProcess string
	Unit            string
	Quantity        decimal.Decimal `gorm:"type:numeric"`
	MaterialName    string
	Node            gormNode `gorm:"foreignKey:NodeDbId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	NodeDbId        uint64   `gorm:"primaryKey"`
}

func (r *assetRepositoryGorm) SaveAsset(ctx context.Context, transactionId TransactionId, iAsset *model.Asset) error {
	err := r.nodeRepository.UpsertNode(ctx, transactionId, &iAsset.Node)
	if err != nil {
		return err
	}

	tx, err := r.transactionManagerGorm.GetTransaction(ctx, transactionId)
	if err != nil {
		return err
	}

	asset := gormAsset{
		CreationProcess: iAsset.CreationProcess,
		Unit:            iAsset.Unit,
		Quantity:        iAsset.Quantity,
		MaterialName:    iAsset.MaterialName,
		NodeDbId:        uint64(iAsset.DbId),
	}
	err = tx.Omit(clause.Associations).Create(&asset).Error

	iAsset.Node.DbId = model.DbId(asset.Node.ID)
	return err
}

func (r *assetRepositoryGorm) fetchAssetsFromNodes(ctx context.Context, tx *gorm.DB, nodes []model.Node) ([]model.Asset, error) {
	assets := []gormAsset{}

	nodesIds := make([]model.DbId, 0, len(nodes))
	nodeIdMap := map[model.DbId]model.Node{}

	for i := range nodes {
		nodesIds = append(nodesIds, nodes[i].DbId)
		nodeIdMap[nodes[i].DbId] = nodes[i]
	}

	err := tx.Where("node_id IN ?", nodesIds).Omit("Node").Order("node_id asc").Find(&assets).Error
	if err != nil {
		return []model.Asset{}, err
	}

	ret := []model.Asset{}
	for i := range assets {
		nodeDbId := model.DbId(assets[i].NodeDbId)
		asset := model.Asset{
			Node:            nodeIdMap[nodeDbId],
			CreationProcess: assets[i].CreationProcess,
			Unit:            assets[i].Unit,
			Quantity:        assets[i].Quantity,
			MaterialName:    assets[i].MaterialName,
		}

		ret = append(ret, asset)
	}

	return ret, nil
}

func (r *assetRepositoryGorm) FetchAssetsByOwner(ctx context.Context, transactionId TransactionId, namespace string, ownerPublicKey string) ([]model.Asset, error) {
	tx, err := r.transactionManagerGorm.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model.Asset{}, err
	}

	nodes, err := r.nodeRepository.FetchNodesByOwnerPublicKey(ctx, transactionId, ownerPublicKey)

	ret, err := r.fetchAssetsFromNodes(ctx, tx, nodes)
	if err != nil {
		return []model.Asset{}, err
	}

	return ret, nil
}

func (r *assetRepositoryGorm) FetchAssetsByIds(ctx context.Context, transactionId TransactionId, namespace string, iIds map[model.NodeId]bool) ([]model.Asset, error) {
	tx, err := r.transactionManagerGorm.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model.Asset{}, err
	}

	nodes, err := r.nodeRepository.FetchNodesByNodeId(ctx, transactionId, namespace, iIds)
	if err != nil {
		return []model.Asset{}, err
	}

	ret, err := r.fetchAssetsFromNodes(ctx, tx, nodes)
	if err != nil {
		return []model.Asset{}, err
	}

	return ret, nil
}
