package repository_server

import (
	"context"
	"sig_graph_scp/pkg/model"
	model_server "sig_graph_scp/pkg/server/model"

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
	ID             uint64 `gorm:"primaryKey,autoIncrement"`
	NodeID         string `gorm:"column:node_id;index:node_id_and_namespace,unique"`
	Namespace      string `gorm:"column:node_namespace;index:node_id_and_namespace,unique"`
	NodeType       string `gorm:"not null"`
	IsFinalized    bool   `gorm:"not null"`
	CreatedTime    uint64 `gorm:"not null"`
	UpdatedTime    uint64 `gorm:"not null"`
	Signature      string `gorm:"column:node_signature;not null"`
	OwnerPublicKey string `gorm:"not null"`

	PublicEdges  []gormPublicEdge  `gorm:"foreignKey:NodeDbId"`
	PrivateEdges []gormPrivateEdge `gorm:"foreignKey:NodeDbId"`
}

type gormPublicEdge struct {
	NodeDbId            uint64 `gorm:"primaryKey,priority:1"`
	Value               string `gorm:"column:other_node_id;primaryKey,priority:2"`
	IsThisNodeTheParent bool
}

type gormPrivateEdge struct {
	NodeDbId            uint64 `gorm:"primaryKey,priority:1"`
	IsThisNodeTheParent bool

	ThisId     model_server.NodeId `gorm:"column:this_node_id"`
	ThisHash   string              `gorm:"primaryKey,priority:2"`
	ThisSecret string

	OtherId     model_server.NodeId `gorm:"column:other_node_id"`
	OtherHash   string
	OtherSecret string
}

func (r *nodeRepositoryGorm) UpsertNode(
	ctx context.Context,
	transactionId TransactionId,
	iNode *model_server.Node,
) error {
	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return err
	}

	gormNode := gormNode{
		NodeID:         string(iNode.Id),
		Namespace:      iNode.Namespace,
		NodeType:       string(iNode.NodeType),
		IsFinalized:    iNode.IsFinalized,
		CreatedTime:    iNode.CreatedTime,
		UpdatedTime:    iNode.UpdatedTime,
		Signature:      iNode.Signature,
		OwnerPublicKey: iNode.OwnerPublicKey,
	}

	for publicParent := range iNode.PublicParentsIds {
		id := gormPublicEdge{
			Value:               publicParent,
			IsThisNodeTheParent: false,
		}
		gormNode.PublicEdges = append(gormNode.PublicEdges, id)
	}

	for publicChildren := range iNode.PublicChildrenIds {
		id := gormPublicEdge{
			Value:               publicChildren,
			IsThisNodeTheParent: true,
		}
		gormNode.PublicEdges = append(gormNode.PublicEdges, id)
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
	).Omit(clause.Associations).Create(&gormNode).Error
	if err != nil {
		return err
	}

	for thisHash, privateParent := range iNode.PrivateParentsIds {
		id := gormPrivateEdge{
			NodeDbId:            gormNode.ID,
			IsThisNodeTheParent: false,

			ThisId:     privateParent.ThisId,
			ThisHash:   thisHash,
			ThisSecret: privateParent.ThisSecret,

			OtherId:     privateParent.OtherId,
			OtherHash:   privateParent.OtherHash,
			OtherSecret: privateParent.OtherSecret,
		}
		gormNode.PrivateEdges = append(gormNode.PrivateEdges, id)
	}

	for thisHash, privateChildren := range iNode.PrivateChildrenIds {
		id := gormPrivateEdge{
			NodeDbId:            gormNode.ID,
			IsThisNodeTheParent: true,

			ThisId:     privateChildren.ThisId,
			ThisHash:   thisHash,
			ThisSecret: privateChildren.ThisSecret,

			OtherId:     privateChildren.OtherId,
			OtherHash:   privateChildren.OtherHash,
			OtherSecret: privateChildren.OtherSecret,
		}
		gormNode.PrivateEdges = append(gormNode.PrivateEdges, id)
		if err != nil {
			return err
		}
	}

	// remove unknown private edges so that we can insert new ones
	newHashes := make([]string, 0, len(gormNode.PrivateEdges))
	for i := range gormNode.PrivateEdges {
		newHashes = append(newHashes, gormNode.PrivateEdges[i].ThisHash)
	}
	err = tx.Where("node_db_id = ? AND this_hash IN ? AND (this_node_id='' OR this_hash='')", gormNode.ID, newHashes).Delete(&gormPrivateEdge{}).Error
	if err != nil {
		return err
	}

	if len(gormNode.PrivateEdges) != 0 {
		err = tx.Clauses(
			clause.OnConflict{
				Columns: []clause.Column{
					{Name: "node_db_id"},
					{Name: "this_hash"},
				},
				DoNothing: true,
			},
		).Create(gormNode.PrivateEdges).Error
	}

	gormNode.PrivateEdges = []gormPrivateEdge{}
	err = tx.Model(&gormNode).Association("PrivateEdges").Find(&gormNode.PrivateEdges)
	if err != nil {
		return err
	}

	// upsert public edges
	err = r.upsertPublicEdges(ctx, tx, iNode, &gormNode)
	if err != nil {
		return err
	}

	*iNode = gormNodeToModelNode(&gormNode)

	return nil
}

func (r *nodeRepositoryGorm) upsertPublicEdges(
	ctx context.Context,
	tx *gorm.DB,
	node *model_server.Node,
	gormNode *gormNode,
) error {
	gormNode.PublicEdges = []gormPublicEdge{}
	for thisId := range node.PublicParentsIds {
		id := gormPublicEdge{
			NodeDbId:            gormNode.ID,
			IsThisNodeTheParent: false,
			Value:               thisId,
		}
		gormNode.PublicEdges = append(gormNode.PublicEdges, id)
	}

	for thisId := range node.PublicChildrenIds {
		id := gormPublicEdge{
			NodeDbId:            gormNode.ID,
			IsThisNodeTheParent: true,
			Value:               thisId,
		}
		gormNode.PublicEdges = append(gormNode.PublicEdges, id)
	}

	if len(gormNode.PublicEdges) != 0 {
		err := tx.Clauses(
			clause.OnConflict{
				Columns: []clause.Column{
					{Name: "node_db_id"},
					{Name: "other_node_id"},
				},
				UpdateAll: true,
			},
		).Create(gormNode.PublicEdges).Error

		return err
	}

	return nil
}

func gormNodeToModelNode(node *gormNode) model_server.Node {
	publicParentsIds := map[string]bool{}
	publicChildrenIds := map[string]bool{}
	for _, id := range node.PublicEdges {
		if id.IsThisNodeTheParent {
			publicChildrenIds[id.Value] = true
		} else {
			publicParentsIds[id.Value] = true
		}
	}

	privateParentsIds := map[string]model_server.PrivateId{}
	privateChildrenIds := map[string]model_server.PrivateId{}
	for _, id := range node.PrivateEdges {
		privateId := toModelServePrivateId(&id)
		if id.IsThisNodeTheParent {
			privateChildrenIds[id.ThisHash] = privateId
		} else {
			privateParentsIds[id.ThisHash] = privateId
		}
	}

	modelNode := model_server.Node{
		NodeType:           model.ENodeType(node.NodeType),
		NodeDbId:           model_server.NodeDbId(node.ID),
		Id:                 model_server.NodeId(node.NodeID),
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

func (r *nodeRepositoryGorm) FetchNodesByOwnerPublicKey(
	ctx context.Context,
	transactionId TransactionId,
	nodeType model.ENodeType,
	namespace string,
	ownerPublicKey string,
	finalizeStatus []bool,
	pagination PaginationOption[model_server.NodeDbId],
) ([]model_server.Node, error) {
	nodes := []gormNode{}

	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model_server.Node{}, err
	}

	err = tx.Preload("PublicEdges").Preload("PrivateEdges").Where("node_namespace = ? AND owner_public_key = ? AND node_type = ? AND is_finalized IN ? AND id >= ?", namespace, ownerPublicKey, nodeType, finalizeStatus, pagination.MinId).Limit(int(pagination.Limit)).Order("id asc").Find(&nodes).Error
	if err != nil {
		return []model_server.Node{}, err
	}

	ret := []model_server.Node{}
	for i := range nodes {
		ret = append(ret, gormNodeToModelNode(&nodes[i]))
	}

	return ret, nil
}

func (r *nodeRepositoryGorm) FetchNodesByNodeId(
	ctx context.Context,
	transactionId TransactionId,
	nodeType model.ENodeType,
	namespace string,
	iIds map[model_server.NodeId]bool,
) ([]model_server.Node, error) {
	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model_server.Node{}, err
	}

	ids := []model_server.NodeId{}
	for id := range iIds {
		ids = append(ids, id)
	}

	nodes := []gormNode{}
	err = tx.Preload("PublicEdges").Preload("PrivateEdges").Where("node_namespace = ? AND node_id IN ? AND node_type = ?", namespace, ids, nodeType).Find(&nodes).Error
	if err != nil {
		return []model_server.Node{}, err
	}

	ret := []model_server.Node{}
	for i := range nodes {
		ret = append(ret, gormNodeToModelNode(&nodes[i]))
	}

	return ret, nil
}

func (r *nodeRepositoryGorm) FetchNodesByDbId(
	ctx context.Context,
	transactionId TransactionId,
	nodeType model.ENodeType,
	namespace string,
	iIds map[model_server.NodeDbId]bool,
) ([]model_server.Node, error) {
	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return []model_server.Node{}, err
	}

	ids := []model_server.NodeDbId{}
	for id := range iIds {
		ids = append(ids, id)
	}

	nodes := []gormNode{}
	err = tx.Preload("PublicEdges").Preload("PrivateEdges").Where("node_namespace = ? AND id IN ? AND node_type = ?", namespace, ids, nodeType).Find(&nodes).Error
	if err != nil {
		return []model_server.Node{}, err
	}

	ret := []model_server.Node{}
	for i := range nodes {
		ret = append(ret, gormNodeToModelNode(&nodes[i]))
	}

	return ret, nil
}

func (r *nodeRepositoryGorm) FetchPrivateEdgesByNodeIds(
	ctx context.Context,
	transactionId TransactionId,
	namespace string,
	edges []EdgeNodeId,
) ([]model_server.PrivateId, error) {
	tx, err := r.transactionManager.GetTransaction(ctx, transactionId)
	if err != nil {
		return nil, err
	}

	edgeTuple := make([][2]string, 0, len(edges))
	for i := range edges {
		edgeTuple = append(edgeTuple, [2]string{
			string(edges[i].Parent),
			string(edges[i].Child),
		})
	}

	gormPrivateEdges := []gormPrivateEdge{}
	err = tx.Raw(`
		SELECT 
			e.node_db_id,
			e.this_node_id, 
			e.this_secret, 
			e.this_hash,
			e.other_node_id, 
			e.other_secret, 
			e.other_hash
		FROM gorm_private_edges e
		JOIN gorm_nodes n
			ON n.node_namespace = ?
			AND e.node_db_id = n.id
		WHERE (e.this_node_id, e.other_node_id) IN ?
	`, namespace, edgeTuple).Scan(&gormPrivateEdges).Error
	if err != nil {
		return nil, err
	}

	privateEdges := make([]model_server.PrivateId, 0, len(gormPrivateEdges))
	for i := range gormPrivateEdges {
		privateEdges = append(privateEdges, toModelServePrivateId(&gormPrivateEdges[i]))
	}
	return privateEdges, nil
}

func toModelServePrivateId(
	privateId *gormPrivateEdge,
) model_server.PrivateId {
	return model_server.PrivateId{
		ThisId:     privateId.ThisId,
		ThisSecret: privateId.ThisSecret,
		ThisHash:   privateId.ThisHash,

		OtherId:     privateId.OtherId,
		OtherSecret: privateId.OtherSecret,
		OtherHash:   privateId.OtherHash,
	}
}
