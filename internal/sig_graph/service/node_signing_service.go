package service_sig_graph

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	model_sig_graph "sig_graph_scp/pkg/sig_graph/model"
	"sig_graph_scp/pkg/utility"
)

type nodeSigningService struct {
}

func NewNodeSigningService() *nodeSigningService {
	return &nodeSigningService{}
}

func (s *nodeSigningService) Sign(ctx context.Context, userKeyPair *model_sig_graph.UserKeyPair, node any) (string, error) {
	nodeMap := map[string]any{}
	nodeJson, err := json.Marshal(node)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(nodeJson, &nodeMap)
	if err != nil {
		return "", err
	}

	delete(nodeMap, "signature")
	nodeWithoutSignatureJson, err := json.Marshal(nodeMap)
	if err != nil {
		return "", err
	}

	signature, err := s.sign(string(nodeWithoutSignatureJson), userKeyPair.Private)
	if err != nil {
		return "", err
	}

	base64Signature := base64.StdEncoding.EncodeToString([]byte(signature))
	return base64Signature, nil
}

func (s *nodeSigningService) sign(data string, privateKey string) (string, error) {
	block, _ := pem.Decode([]byte(privateKey))
	privateKeyParsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	hash := sha512.Sum512([]byte(data))

	if rsaKey, ok := privateKeyParsed.(*rsa.PrivateKey); ok {
		signature, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA512, hash[:])
		if err != nil {
			return "", err
		}

		return string(signature), nil
	} else if ecdsaKey, ok := privateKeyParsed.(*ecdsa.PrivateKey); ok {
		signature, err := ecdsa.SignASN1(rand.Reader, ecdsaKey, hash[:])
		if err != nil {
			return "", err
		}

		return string(signature), nil
	} else {
		return "", fmt.Errorf("%w: unsupported signature algorithm", utility.ErrInvalidArgument)
	}
}
