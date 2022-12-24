package model_asset_transfer

type AcceptAssetEvent struct {
	AckId      string
	IsAccepted bool
	Message    string
}
