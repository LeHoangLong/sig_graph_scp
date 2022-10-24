package service_asset_transfer

import (
	"context"
	"crypto/rand"
	"math/big"
)

type secretIdGeneratorCrypto struct {
	length uint32
}

func NewSecretIdGeneratorCrypto(length uint32) *secretIdGeneratorCrypto {
	return &secretIdGeneratorCrypto{
		length: length,
	}
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var max = big.NewInt(int64(len(letters)))

func (s *secretIdGeneratorCrypto) NewSecretId(ctx context.Context) (string, error) {
	ret := make([]byte, s.length)
	for i := 0; i < int(s.length); i++ {
		randomIdx, _ := rand.Int(rand.Reader, max)
		ret[i] = letters[randomIdx.Int64()]
	}

	return string(ret), nil
}
