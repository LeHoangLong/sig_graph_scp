package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func main() {
	curve := elliptic.P521()
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}

	{
		file, err := os.Create("pub.pem")
		if err != nil {
			panic(err)
		}
		data, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			panic(err)
		}
		block := &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: data,
		}
		err = pem.Encode(file, block)
		if err != nil {
			panic(err)
		}
	}

	{
		file, err := os.Create("key.pem")
		if err != nil {
			panic(err)
		}
		data, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			panic(err)
		}
		block := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: data,
		}
		err = pem.Encode(file, block)
		if err != nil {
			panic(err)
		}
	}
}
