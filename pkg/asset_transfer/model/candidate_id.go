package model_asset_transfer

type CandidateId struct {
	Id        string `json:"id"`
	Secret    string `json:"secret"`
	Signature string `json:"signature"`
}
