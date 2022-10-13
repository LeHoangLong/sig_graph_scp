package repository_server

import (
	"context"
	"sig_graph_scp/pkg/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type nodeRepositoryGorm struct {
	transactionManager *transactionManagerGorm
}

var _ (NodeRepositoryI) = (*nodeRepositoryGorm)(nil)

func NewNodeRepositoryGorm(
	transactionManager *transactionManagerGorm,
) *nodeRepositoryGorm {
	return &nodeRepositoryGorm{
		transactionManager: transactionManager,
	}
}

type gormNode struct {
	gorm.Model
	ID             uint64 `gorm:"primaryKey,autoIncrement"`
	NodeID         string `gorm:"index:node_id_and_namespace,unique"`
	Namespace      string `gorm:"index:node_id_and_namespace,unique"`
	IsFinalized    bool   `gorm:"not null"`
	CreatedTime    uint64 `gorm:"not null"`
	UpdatedTime    uint64 `gorm:"not null"`
	Signature      string `gorm:"not null"`
	OwnerPublicKey string `gorm:"not null"`

	PublicParentsIds  []publicEdge `gorm:"foreignKey:NodeDbId"`
	PublicChildrenIds []publicEdge `gorm:"foreignKey:NodeDbId"`

	PrivateParentsIds  []privateEdge `gorm:"foreignKey:NodeDbId"`
	PrivateChildrenIds []privateEdge `gorm:"foreignKey:NodeDbId"`
}

type publicEdge struct {
	gorm.Model
	NodeDbId uint64 `gorm:"primaryKey,priority:1"`
	Value    string `gorm:"primaryKey,priority:2"`
}

type privateEdge struct {
	gorm.Model
	NodeDbId uint64 `gorm:"primaryKey,priority:1"`
	Value    string `gorm:"primaryKey,priority:2"`
	Secret   string `gorm:"not null;default:"`
}

func (r *nodeRepositoryGorm) UpsertNode(
	ctx context.Context,
	transactionId TransactionId,
	iNode *model.Node,
) error {
	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return err
	}

	gormNode := gormNode{
		NodeID:         string(iNode.Id),
		Namespace:      iNode.Namespace,
		IsFinalized:    iNode.IsFinalized,
		CreatedTime:    iNode.CreatedTime,
		UpdatedTime:    iNode.UpdatedTime,
		Signature:      iNode.Signature,
		OwnerPublicKey: iNode.OwnerPublicKey,
	}

	for publicParent := range iNode.PublicParentsIds {
		id := publicEdge{
			Value: publicParent,
		}
		gormNode.PublicParentsIds = append(gormNode.PublicParentsIds, id)
	}

	for publicChildren := range iNode.PublicChildrenIds {
		id := publicEdge{
			Value: publicChildren,
		}
		gormNode.PublicParentsIds = append(gormNode.PublicParentsIds, id)
	}

	for privateParent := range iNode.PrivateParentsIds {
		id := privateEdge{
			Value:  string(privateParent.Id),
			Secret: privateParent.Secret,
		}
		gormNode.PrivateParentsIds = append(gormNode.PrivateParentsIds, id)
	}

	for privateChildren := range iNode.PrivateChildrenIds {
		id := privateEdge{
			Value:  string(privateChildren.Id),
			Secret: privateChildren.Secret,
		}
		gormNode.PrivateChildrenIds = append(gormNode.PrivateChildrenIds, id)
	}

	err = tx.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: "node_id"},
				{Name: "namespace"},
			},
			UpdateAll: true,
		},
	).Clauses(
		clause.Returning{
			Columns: []clause.Column{
				{Name: "id"},
			},
		},
	).Create(&gormNode).Error

	iNode.DbId = model.DbId(gormNode.ID)
	return err
}

func gormNodeToModelNode(node gormNode) model.Node {
	publicParentsIds := map[string]bool{}
	for _, id := range node.PublicParentsIds {
		publicParentsIds[id.Value] = true
	}

	publicChildrenIds := map[string]bool{}
	for _, id := range node.PublicChildrenIds {
		publicChildrenIds[id.Value] = true
	}

	privateParentsIds := map[model.PrivateId]bool{}
	for _, id := range node.PrivateParentsIds {
		privateId := model.PrivateId{
			Id:     model.NodeId(id.Value),
			Secret: id.Secret,
		}
		privateParentsIds[privateId] = true
	}

	privateChildrenIds := map[model.PrivateId]bool{}
	for _, id := range node.PrivateChildrenIds {
		privateId := model.PrivateId{
			Id:     model.NodeId(id.Value),
			Secret: id.Secret,
		}
		privateChildrenIds[privateId] = true
	}

	modelNode := model.Node{
		DbId:               model.DbId(node.ID),
		Id:                 model.NodeId(node.NodeID),
		Namespace:          node.Namespace,
		PublicParentsIds:   publicParentsIds,
		PublicChildrenIds:  publicChildrenIds,
		PrivateParentsIds:  privateParentsIds,
		PrivateChildrenIds: privateChildrenIds,
		IsFinalized:        node.IsFinalized,
		CreatedTime:        node.CreatedTime,
		UpdatedTime:        node.UpdatedTime,
		Signature:          node.Signature,
		OwnerPublicKey:     node.OwnerPublicKey,
	}
	return modelNode
}

func (r *nodeRepositoryGorm) FetchNodesByOwnerPublicKey(ctx context.Context, transactionId TransactionId, ownerPublicKey string) ([]model.Node, error) {
	nodes := []gormNode{}

	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model.Node{}, err
	}

	tx.Where("owner_public_key = ?", ownerPublicKey).Order("id asc").Find(&nodes)

	ret := []model.Node{}
	for i := range nodes {
		ret = append(ret, gormNodeToModelNode(nodes[i]))
	}

	return ret, nil
}

func (r *nodeRepositoryGorm) FetchNodesByNodeId(
	ctx context.Context,
	transactionId TransactionId,
	namespace string,
	iIds map[model.NodeId]bool,
) ([]model.Node, error) {
	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model.Node{}, err
	}

	ids := []model.NodeId{}
	for id := range iIds {
		ids = append(ids, id)
	}

	nodes := []gormNode{}
	err = tx.Where("namespace = ? AND node_id IN ?", namespace, iIds).Find(&nodes).Error
	if err != nil {
		return []model.Node{}, err
	}

	ret := []model.Node{}
	for i := range nodes {
		ret = append(ret, gormNodeToModelNode(nodes[i]))
	}

	return ret, nil
}
