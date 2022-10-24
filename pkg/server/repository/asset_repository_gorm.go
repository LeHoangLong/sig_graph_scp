package repository_server

import (
	"context"
	model_server "sig_graph_scp/pkg/server/model"

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
	CreationProcess string
	Unit            string
	Quantity        decimal.Decimal `gorm:"type:numeric"`
	MaterialName    string
	Node            gormNode `gorm:"foreignKey:NodeDbId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	NodeDbId        uint64   `gorm:"primaryKey"`
}

func (r *assetRepositoryGorm) SaveAsset(ctx context.Context, transactionId TransactionId, iAsset *model_server.Asset) error {
	iAsset.Node.NodeType = "asset"
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
	return err
}

func (r *assetRepositoryGorm) fetchAssetsFromNodes(ctx context.Context, tx *gorm.DB, nodes []model_server.Node) ([]model_server.Asset, error) {
	assets := []gormAsset{}

	nodesIds := make([]model_server.DbId, 0, len(nodes))
	nodeIdMap := map[model_server.DbId]model_server.Node{}

	for i := range nodes {
		nodesIds = append(nodesIds, nodes[i].DbId)
		nodeIdMap[nodes[i].DbId] = nodes[i]
	}

	err := tx.Where("node_db_id IN ?", nodesIds).Omit("Node").Order("node_db_id asc").Find(&assets).Error
	if err != nil {
		return []model_server.Asset{}, err
	}

	ret := []model_server.Asset{}
	for i := range assets {
		nodeDbId := model_server.DbId(assets[i].NodeDbId)
		asset := model_server.Asset{
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

func (r *assetRepositoryGorm) FetchAssetsByOwner(ctx context.Context, transactionId TransactionId, namespace string, ownerPublicKey string) ([]model_server.Asset, error) {
	tx, err := r.transactionManagerGorm.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model_server.Asset{}, err
	}

	nodes, err := r.nodeRepository.FetchNodesByOwnerPublicKey(ctx, transactionId, "asset", ownerPublicKey)

	ret, err := r.fetchAssetsFromNodes(ctx, tx, nodes)
	if err != nil {
		return []model_server.Asset{}, err
	}

	return ret, nil
}

func (r *assetRepositoryGorm) FetchAssetsByIds(ctx context.Context, transactionId TransactionId, namespace string, iIds map[model_server.NodeId]bool) ([]model_server.Asset, error) {
	tx, err := r.transactionManagerGorm.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model_server.Asset{}, err
	}

	nodes, err := r.nodeRepository.FetchNodesByNodeId(ctx, transactionId, "asset", namespace, iIds)
	if err != nil {
		return []model_server.Asset{}, err
	}

	ret, err := r.fetchAssetsFromNodes(ctx, tx, nodes)
	if err != nil {
		return []model_server.Asset{}, err
	}

	return ret, nil
}
