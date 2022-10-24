package model_sig_graph

import "sig_graph_scp/pkg/model"

type Node struct {
	Id                       string          `json:"id"`
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
	id string,
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
