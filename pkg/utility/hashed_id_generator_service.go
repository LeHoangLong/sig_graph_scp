package utility

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
)

type hashedIdGeneratorService struct {
}

func NewHashedIdGeneratorService() *hashedIdGeneratorService {
	return &hashedIdGeneratorService{}
}

func (s *hashedIdGeneratorService) GenerateHashedId(ctx context.Context, id string, secret string) (string, error) {
	secretId := fmt.Sprintf("%s%s", id, secret)
	hashByte := sha512.Sum512([]byte(secretId))
	hash := base64.StdEncoding.EncodeToString(hashByte[:])
	return hash, nil
}
