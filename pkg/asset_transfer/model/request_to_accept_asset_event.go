package model_asset_transfer

type RequestToAcceptAssetEvent struct {
	TimeMs                    uint64               `json:"time_ms"`
	AckId                     string               `json:"ack_id"`
	AssetId                   string               `json:"asset_id"`
	PeerPemPublicKey          string               `json:"peer_pem_public_key"`
	UserPemPublicKey          string               `json:"user_id"`
	ExposedPrivateConnections map[string]PrivateId `json:"exposed_private_connections"`
	Candidates                []CandidateId        `json:"candidates"`
}
