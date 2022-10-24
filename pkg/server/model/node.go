package model_server

import (
	"sig_graph_scp/pkg/model"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
)

type DbId uint64

type NodeId string

type PrivateId struct {
	ThisId      NodeId `json:"this_id"`
	ThisHash    string `json:"this_hash"`
	ThisSecret  string `json:"this_secret"`
	OtherId     NodeId `json:"other_id"`
	OtherHash   string `json:"other_hash"`
	OtherSecret string `json:"other_secret"`
}

type Node struct {
	DbId DbId `json:"db_id"`

	Id        NodeId `json:"id"`
	Namespace string `json:"-"`
	NodeType  string `json:"node_type"`

	PublicParentsIds  map[string]bool `json:"public_parents_ids"`
	PublicChildrenIds map[string]bool `json:"public_children_ids"`

	PrivateParentsIds  map[string]PrivateId `json:"private_parents_ids"` // key is the hash
	PrivateChildrenIds map[string]PrivateId `json:"private_children_ids"`

	IsFinalized    bool   `json:"is_finalized"`
	CreatedTime    uint64 `json:"created_time"`
	UpdatedTime    uint64 `json:"updated_time"`
	Signature      string `json:"signature"`
	OwnerPublicKey string `json:"owner_public_key"`
}

func FromSigGraphNode(node *model_sig_graph.Node, dbId DbId, namespace string, secretParentIds map[string]PrivateId, secretChildrenIds map[string]PrivateId) Node {
	return Node{
		DbId:               dbId,
		Id:                 NodeId(node.Id),
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

func ToSigGraphNode(node *Node) model_sig_graph.Node {
	privateParentsHashedIds := map[string]bool{}
	for hash := range node.PrivateParentsIds {
		privateParentsHashedIds[hash] = true
	}

	privateChildrenHashedIds := map[string]bool{}
	for hash := range node.PrivateChildrenIds {
		privateChildrenHashedIds[hash] = true
	}

	return model_sig_graph.Node{
		Id:                       string(node.Id),
		PublicParentsIds:         node.PublicParentsIds,
		PublicChildrenIds:        node.PublicChildrenIds,
		PrivateParentsHashedIds:  privateParentsHashedIds,
		PrivateChildrenHashedIds: privateChildrenHashedIds,
		IsFinalized:              node.IsFinalized,
		NodeType:                 model.ENodeType(node.NodeType),
		CreatedTime:              node.CreatedTime,
		UpdatedTime:              node.UpdatedTime,
		Signature:                node.Signature,
		OwnerPublicKey:           node.OwnerPublicKey,
	}
}
