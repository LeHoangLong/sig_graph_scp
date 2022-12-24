package repository_server

import (
	"context"
	"sig_graph_scp/pkg/model"
	model_server "sig_graph_scp/pkg/server/model"
	"sig_graph_scp/pkg/utility"

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
	iAsset.Node.NodeType = model.ENodeTypeAsset
	err := r.nodeRepository.UpsertNode(ctx, transactionId, &iAsset.Node)
	if err != nil {
		return err
	}
	iAsset.Node.Extra = iAsset

	tx, err := r.transactionManagerGorm.GetTransaction(ctx, transactionId)
	if err != nil {
		return err
	}

	asset := gormAsset{
		CreationProcess: iAsset.CreationProcess,
		Unit:            iAsset.Unit,
		Quantity:        iAsset.Quantity,
		MaterialName:    iAsset.MaterialName,
		NodeDbId:        uint64(iAsset.NodeDbId),
	}
	err = tx.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: "node_db_id"},
			},
			UpdateAll: true,
		},
	).Omit(clause.Associations).Create(&asset).Error
	return err
}

func (r *assetRepositoryGorm) fetchAssetsFromNodes(ctx context.Context, tx *gorm.DB, nodes []model_server.Node) ([]model_server.Asset, error) {
	assets := []gormAsset{}

	nodesIds := make([]model_server.NodeDbId, 0, len(nodes))
	nodeIdMap := map[model_server.NodeDbId]model_server.Node{}

	for i := range nodes {
		nodesIds = append(nodesIds, nodes[i].NodeDbId)
		nodeIdMap[nodes[i].NodeDbId] = nodes[i]
	}

	err := tx.Where("node_db_id IN ?", nodesIds).Omit("Node").Order("node_db_id asc").Find(&assets).Error
	if err != nil {
		return []model_server.Asset{}, err
	}

	ret := []model_server.Asset{}
	for i := range assets {
		nodeDbId := model_server.NodeDbId(assets[i].NodeDbId)
		asset := model_server.Asset{
			Node:            nodeIdMap[nodeDbId],
			CreationProcess: assets[i].CreationProcess,
			Unit:            assets[i].Unit,
			Quantity:        assets[i].Quantity,
			MaterialName:    assets[i].MaterialName,
		}

		asset.Node.Extra = &asset
		ret = append(ret, asset)
	}

	return ret, nil
}

func (r *assetRepositoryGorm) FetchAssetsByOwner(
	ctx context.Context,
	transactionId TransactionId,
	namespace string,
	ownerPublicKey string,
	finalizeStatus []bool,
	pagination PaginationOption[model_server.NodeDbId],
) ([]model_server.Asset, error) {
	tx, err := r.transactionManagerGorm.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model_server.Asset{}, err
	}

	nodes, err := r.nodeRepository.FetchNodesByOwnerPublicKey(
		ctx,
		transactionId,
		model.ENodeTypeAsset,
		namespace,
		ownerPublicKey,
		finalizeStatus,
		pagination,
	)

	ret, err := r.fetchAssetsFromNodes(ctx, tx, nodes)
	if err != nil {
		return []model_server.Asset{}, err
	}

	return ret, nil
}

func (r *assetRepositoryGorm) FetchAssetsByIds(ctx context.Context, transactionId TransactionId, namespace string, ids map[model_server.NodeId]bool) ([]model_server.Asset, error) {
	tx, err := r.transactionManagerGorm.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model_server.Asset{}, err
	}

	nodes, err := r.nodeRepository.FetchNodesByNodeId(ctx, transactionId, model.ENodeTypeAsset, namespace, ids)
	if err != nil {
		return []model_server.Asset{}, err
	}

	ret, err := r.fetchAssetsFromNodes(ctx, tx, nodes)
	if err != nil {
		return []model_server.Asset{}, err
	}

	return ret, nil
}

func (r *assetRepositoryGorm) FetchAssetsByDbIds(ctx context.Context, transactionId TransactionId, namespace string, ids map[model_server.NodeDbId]bool) ([]model_server.Asset, error) {
	tx, err := r.transactionManagerGorm.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model_server.Asset{}, err
	}

	nodes, err := r.nodeRepository.FetchNodesByDbId(ctx, transactionId, model.ENodeTypeAsset, namespace, ids)
	if err != nil {
		return []model_server.Asset{}, err
	}

	ret, err := r.fetchAssetsFromNodes(ctx, tx, nodes)
	if err != nil {
		return []model_server.Asset{}, err
	}

	return ret, nil
}

func (r *assetRepositoryGorm) FetchNode(
	ctx context.Context,
	txId TransactionId,
	node *model_server.Node,
) (model_server.Node, error) {
	tx, err := r.transactionManagerGorm.GetTransaction(ctx, txId)
	if err != nil {
		return model_server.Node{}, err
	}

	assets, err := r.fetchAssetsFromNodes(ctx, tx, []model_server.Node{
		*node,
	})

	if err != nil {
		return model_server.Node{}, err
	}

	if len(assets) == 0 {
		return model_server.Node{}, utility.ErrNotFound
	}

	return assets[0].Node, nil
}

func (r *assetRepositoryGorm) UpsertNode(
	ctx context.Context,
	txId TransactionId,
	node *model_server.Node,
) (model_server.Node, error) {
	if asset, ok := node.Extra.(*model_server.Asset); ok {
		clonedAsset := *asset
		err := r.SaveAsset(
			ctx,
			txId,
			&clonedAsset,
		)
		if err != nil {
			return model_server.Node{}, err
		}

		return clonedAsset.Node, nil
	}

	return model_server.Node{}, utility.ErrInvalidArgument
}
