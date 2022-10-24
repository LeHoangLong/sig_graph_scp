package model_asset_transfer

type PrivateId struct {
	ThisId      string `json:"this_id"`
	ThisHash    string `json:"this_hash"`
	ThisSecret  string `json:"this_secret"`
	OtherId     string `json:"other_id"`
	OtherHash   string `json:"other_hash"`
	OtherSecret string `json:"other_secret"`
}
