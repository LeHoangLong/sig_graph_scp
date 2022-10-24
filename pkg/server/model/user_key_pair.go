package model_server

type UserKeyPairId = uint64

// public and private key in pem encoded form
type UserKeyPair struct {
	Id      UserKeyPairId
	Public  string
	Private string
}
