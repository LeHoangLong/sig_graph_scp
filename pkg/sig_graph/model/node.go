package model_sig_graph

import "sig_graph_scp/pkg/model"

type Node struct {
	Id                       model.NodeId    `json:"id"`
	PublicParentsIds         map[string]bool `json:"public_parents_ids"`
	PublicChildrenIds        map[string]bool `json:"public_children_ids"`
	PrivateParentsHashedIds  map[string]bool `json:"private_parents_hashed_ids"`
	PrivateChildrenHashedIds map[string]bool `json:"private_children_hashed_ids"`
	IsFinalized              bool            `json:"is_finalized"`
	NodeType                 model.ENodeType `json:"type"`
	CreatedTime              uint64          `json:"created_time"`
	UpdatedTime              uint64          `json:"updated_time"`
	Signature                string          `json:"signature"`
	OwnerPublicKey           string          `json:"owner_public_key"`
}

func NewDefaultNode(
	id model.NodeId,
	nodeType model.ENodeType,
	createdTime uint64,
	updatedTime uint64,
	signature string,
	ownerPublicKey string,
) Node {
	return Node{
		Id:                       id,
		PublicParentsIds:         map[string]bool{},
		PublicChildrenIds:        map[string]bool{},
		PrivateParentsHashedIds:  map[string]bool{},
		PrivateChildrenHashedIds: map[string]bool{},
		IsFinalized:              false,
		NodeType:                 nodeType,
		CreatedTime:              createdTime,
		UpdatedTime:              updatedTime,
		Signature:                signature,
		OwnerPublicKey:           ownerPublicKey,
	}
}

func ToModelNode(node *Node, dbId model.DbId, namespace string, secretParentIds map[model.PrivateId]bool, secretChildrenIds map[model.PrivateId]bool) model.Node {
	return model.Node{
		DbId:               dbId,
		Id:                 node.Id,
		Namespace:          namespace,
		PublicParentsIds:   node.PublicParentsIds,
		PublicChildrenIds:  node.PublicChildrenIds,
		PrivateParentsIds:  secretParentIds,
		PrivateChildrenIds: secretChildrenIds,
		IsFinalized:        node.IsFinalized,
		CreatedTime:        node.CreatedTime,
		UpdatedTime:        node.UpdatedTime,
		Signature:          node.Signature,
		OwnerPublicKey:     node.OwnerPublicKey,
	}
}
