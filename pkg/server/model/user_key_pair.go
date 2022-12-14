package model_server

import model_sig_graph "sig_graph_scp/pkg/sig_graph/model"

type UserKeyPairId = uint64

// public and private key in pem encoded form
type UserKeyPair struct {
	Id      UserKeyPairId `json:"id"`
	Public  string        `json:"public"`
	Private string        `json:"private"`
}

func ToSigGraphUserKeyPair(
	keyPair *UserKeyPair,
) model_sig_graph.UserKeyPair {
	return model_sig_graph.UserKeyPair{
		Public:  keyPair.Public,
		Private: keyPair.Private,
	}
}
