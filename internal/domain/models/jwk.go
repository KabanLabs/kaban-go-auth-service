package models

import (
	"crypto/rsa"
	"encoding/base64"
	"math/big"
)

type JWK struct {
	Kid     string `json:"kid,omitempty"`
	Kty     string `json:"kty,omitempty"`
	Alg     string `json:"alg,omitempty"`
	Use     string `json:"use,omitempty"`
	N       string `json:"n,omitempty"`
	E       string `json:"e,omitempty"`
	Exp     int64  `json:"exp,omitempty"`
	Created int64  `json:"created,omitempty"`
}

func (j *JWK) ToPublicKey() (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(j.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(j.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	var e int
	if len(eBytes) > 0 {
		e = int(new(big.Int).SetBytes(eBytes).Int64())
	} else {
		e = 65537 // дефолтный
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}
