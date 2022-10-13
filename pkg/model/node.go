package model

type DbId uint64

type NodeId string

type PrivateId struct {
	Id     NodeId `json:"id"`
	Secret string `json:"secret"`
}

type ENodeType string

const (
	ENodeTypeAsset ENodeType = "asset"
)

type Node struct {
	DbId DbId `json:"db_id"`

	Id        NodeId `json:"id"`
	Namespace string `json:"-"`

	PublicParentsIds  map[string]bool `json:"public_parents_ids"`
	PublicChildrenIds map[string]bool `json:"public_children_ids"`

	PrivateParentsIds  map[PrivateId]bool `json:"private_parents_ids"`
	PrivateChildrenIds map[PrivateId]bool `json:"private_children_ids"`

	IsFinalized    bool   `json:"is_finalized"`
	CreatedTime    uint64 `json:"created_time"`
	UpdatedTime    uint64 `json:"updated_time"`
	Signature      string `json:"signature"`
	OwnerPublicKey string `json:"owner_public_key"`
}
