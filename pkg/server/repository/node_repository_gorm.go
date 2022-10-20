package repository_server

import (
	"context"
	"sig_graph_scp/pkg/model"

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
	ID             uint64 `gorm:"primaryKey,autoIncrement"`
	NodeID         string `gorm:"column:node_id;index:node_id_and_namespace,unique"`
	Namespace      string `gorm:"column:node_namespace;index:node_id_and_namespace,unique"`
	NodeType       string `gorm:"not null"`
	IsFinalized    bool   `gorm:"not null"`
	CreatedTime    uint64 `gorm:"not null"`
	UpdatedTime    uint64 `gorm:"not null"`
	Signature      string `gorm:"column:node_signature;not null"`
	OwnerPublicKey string `gorm:"not null"`

	PublicParentsIds  []gormPublicEdge `gorm:"foreignKey:NodeDbId"`
	PublicChildrenIds []gormPublicEdge `gorm:"foreignKey:NodeDbId"`

	PrivateParentsIds  []gormPrivateEdge `gorm:"foreignKey:NodeDbId"`
	PrivateChildrenIds []gormPrivateEdge `gorm:"foreignKey:NodeDbId"`
}

type gormPublicEdge struct {
	NodeDbId uint64 `gorm:"primaryKey,priority:1"`
	Value    string `gorm:"column:other_node_id;primaryKey,priority:2"`
}

type gormPrivateEdge struct {
	NodeDbId uint64 `gorm:"primaryKey,priority:1"`
	Hash     string `gorm:"primaryKey,priority:2"`
	Value    string `gorm:"column:other_node_id"`
	Secret   string `gorm:"column:other_node_id_secret;not null;default:"`
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
		NodeType:       iNode.NodeType,
		IsFinalized:    iNode.IsFinalized,
		CreatedTime:    iNode.CreatedTime,
		UpdatedTime:    iNode.UpdatedTime,
		Signature:      iNode.Signature,
		OwnerPublicKey: iNode.OwnerPublicKey,
	}

	for publicParent := range iNode.PublicParentsIds {
		id := gormPublicEdge{
			Value: publicParent,
		}
		gormNode.PublicParentsIds = append(gormNode.PublicParentsIds, id)
	}

	for publicChildren := range iNode.PublicChildrenIds {
		id := gormPublicEdge{
			Value: publicChildren,
		}
		gormNode.PublicParentsIds = append(gormNode.PublicParentsIds, id)
	}

	for _, privateParent := range iNode.PrivateParentsIds {
		id := gormPrivateEdge{
			Hash:   privateParent.Hash,
			Value:  string(privateParent.Id),
			Secret: privateParent.Secret,
		}
		gormNode.PrivateParentsIds = append(gormNode.PrivateParentsIds, id)
	}

	for _, privateChildren := range iNode.PrivateChildrenIds {
		id := gormPrivateEdge{
			Hash:   privateChildren.Hash,
			Value:  string(privateChildren.Id),
			Secret: privateChildren.Secret,
		}
		gormNode.PrivateChildrenIds = append(gormNode.PrivateChildrenIds, id)
	}

	err = tx.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: "node_id"},
				{Name: "node_namespace"},
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

	privateParentsIds := map[string]model.PrivateId{}
	for _, id := range node.PrivateParentsIds {
		privateId := model.PrivateId{
			Id:     model.NodeId(id.Value),
			Hash:   id.Hash,
			Secret: id.Secret,
		}
		privateParentsIds[id.Hash] = privateId
	}

	privateChildrenIds := map[string]model.PrivateId{}
	for _, id := range node.PrivateChildrenIds {
		privateId := model.PrivateId{
			Id:     model.NodeId(id.Value),
			Secret: id.Secret,
			Hash:   id.Hash,
		}
		privateChildrenIds[id.Hash] = privateId
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

func (r *nodeRepositoryGorm) FetchNodesByOwnerPublicKey(ctx context.Context, transactionId TransactionId, nodeType string, ownerPublicKey string) ([]model.Node, error) {
	nodes := []gormNode{}

	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model.Node{}, err
	}

	tx.Where("owner_public_key = ? AND node_type = ?", ownerPublicKey, nodeType).Order("id asc").Find(&nodes)

	ret := []model.Node{}
	for i := range nodes {
		ret = append(ret, gormNodeToModelNode(nodes[i]))
	}

	return ret, nil
}

func (r *nodeRepositoryGorm) FetchNodesByNodeId(
	ctx context.Context,
	transactionId TransactionId,
	nodeType string,
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
	err = tx.Where("node_namespace = ? AND node_id IN ? AND node_type = ?", namespace, ids, nodeType).Find(&nodes).Error
	if err != nil {
		return []model.Node{}, err
	}

	ret := []model.Node{}
	for i := range nodes {
		ret = append(ret, gormNodeToModelNode(nodes[i]))
	}

	return ret, nil
}
