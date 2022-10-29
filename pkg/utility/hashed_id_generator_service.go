package utility

import (
	"crypto/sha512"
	"fmt"
)

type hashedIdGeneratorService struct {
}

func NewHashedIdGeneratorService() *hashedIdGeneratorService {
	return &hashedIdGeneratorService{}
}

func (s *hashedIdGeneratorService) GenerateHashedId(id string, secret string) (string, error) {
	secretId := fmt.Sprintf("%s%s", id, secret)
	hashByte := sha512.Sum512([]byte(secretId))
	hash := string(hashByte[:])
	return hash, nil
}
