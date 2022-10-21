package model

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

type ENodeType string

const (
	ENodeTypeAsset ENodeType = "asset"
)

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
