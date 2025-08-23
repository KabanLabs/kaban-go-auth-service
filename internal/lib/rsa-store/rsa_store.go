package rsa_store

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/VACdotCS/kaban-go-auth-service/internal/domain/models"
)

const (
	privateKeyPath = ".keys/private.pem"
	publicKeyPath  = ".keys/public.pem"
	jwksFilePath   = ".keys/jwks.json"
	defaultKeyTTL  = 24 * time.Hour
)

type Keys struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	JWK        *models.JWK
	Exp        int64
}

var (
	memKeys  *Keys
	memJWKS  = make(map[string]*models.JWK)
	mu       sync.Mutex
	kidCount int
)

// LoadOrGenerateKeys загружает ключи или создаёт новые
func LoadOrGenerateKeys(bits int, ttl time.Duration) (*Keys, error) {
	if ttl == 0 {
		ttl = defaultKeyTTL
	}

	mu.Lock()
	defer mu.Unlock()

	if memKeys != nil {
		return memKeys, nil
	}

	_ = loadJWKSFromDisk()

	keys := &Keys{}

	privData, err := os.ReadFile(privateKeyPath)
	if err == nil {
		block, _ := pem.Decode(privData)
		if block == nil || block.Type != "RSA PRIVATE KEY" {
			return nil, errors.New("invalid private key PEM")
		}

		privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		keys.PrivateKey = privKey
		keys.PublicKey = &privKey.PublicKey

		// Восстановим JWK из memJWKS по последнему kid
		lastKid := fmt.Sprintf("kid-%d", kidCount)
		if jwk, ok := memJWKS[lastKid]; ok {
			keys.JWK = jwk
			keys.Exp = jwk.Exp
		} else {
			keys.JWK = buildJWK(&privKey.PublicKey, ttl)
			keys.Exp = keys.JWK.Exp
			memJWKS[keys.JWK.Kid] = keys.JWK
			saveJWKSToDisk()
		}

		memKeys = keys
		return keys, nil
	}

	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	keys.PrivateKey = privKey
	keys.PublicKey = &privKey.PublicKey
	keys.JWK = buildJWK(&privKey.PublicKey, ttl)
	keys.Exp = keys.JWK.Exp
	memJWKS[keys.JWK.Kid] = keys.JWK

	if err := savePEMKey(privateKeyPath, privKey); err != nil {
		return nil, err
	}
	if err := savePublicPEMKey(publicKeyPath, &privKey.PublicKey); err != nil {
		return nil, err
	}

	err = saveJWKSToDisk()

	if err != nil {
		return nil, err
	}

	memKeys = keys
	return keys, nil
}

// RotateKey создаёт новый ключ и добавляет его в memJWKS
func RotateKey(bits int, ttl time.Duration) (*Keys, error) {
	if ttl == 0 {
		ttl = defaultKeyTTL
	}

	mu.Lock()
	defer mu.Unlock()

	privKey, err := rsa.GenerateKey(rand.Reader, bits)

	if err != nil {
		return nil, err
	}

	keys := &Keys{
		PrivateKey: privKey,
		PublicKey:  &privKey.PublicKey,
		JWK:        buildJWK(&privKey.PublicKey, ttl),
		Exp:        time.Now().Add(ttl).Unix(),
	}

	memKeys = keys
	memJWKS[keys.JWK.Kid] = keys.JWK

	if err := savePEMKey(privateKeyPath, privKey); err != nil {
		return nil, err
	}
	if err := savePublicPEMKey(publicKeyPath, &privKey.PublicKey); err != nil {
		return nil, err
	}

	saveJWKSToDisk()
	return keys, nil
}

// GetJWKByKid возвращает JWK по kid
func GetJWKByKid(kid string) (*models.JWK, error) {
	mu.Lock()
	defer mu.Unlock()

	jwk, ok := memJWKS[kid]
	if !ok {
		return nil, fmt.Errorf("jwk with kid %s not found", kid)
	}

	return jwk, nil
}

// savePEMKey сохраняет private key
func savePEMKey(filename string, key *rsa.PrivateKey) error {
	if err := os.MkdirAll(".keys", 0755); err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	return pem.Encode(file, block)
}

// savePublicPEMKey сохраняет public key
func savePublicPEMKey(filename string, pubKey *rsa.PublicKey) error {
	if err := os.MkdirAll(".keys", 0755); err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	pubBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return err
	}
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}
	return pem.Encode(file, block)
}

// GetLastKeyId возвращает kid последнего ключа
func GetLastKeyId() string {
	mu.Lock()
	defer mu.Unlock()
	return fmt.Sprintf("kid-%d", kidCount)
}

// buildJWK строит JWK и генерирует уникальный kid
func buildJWK(pub *rsa.PublicKey, ttl time.Duration) *models.JWK {
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	eBytes := big.NewInt(int64(pub.E)).Bytes()
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	kidCount++
	return &models.JWK{
		Kid: fmt.Sprintf("kid-%d", kidCount),
		Kty: "RSA",
		Alg: "RS256",
		Use: "sig",
		N:   n,
		E:   e,
		Exp: time.Now().Add(ttl).Unix(),
	}
}

// saveJWKSToDisk сохраняет memJWKS в JSON
func saveJWKSToDisk() error {
	data, err := json.MarshalIndent(memJWKS, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(jwksFilePath, data, 0644)
}

// loadJWKSFromDisk загружает memJWKS из JSON
func loadJWKSFromDisk() error {
	data, err := os.ReadFile(jwksFilePath)
	if err != nil {
		return err
	}
	var jwks map[string]*models.JWK
	if err := json.Unmarshal(data, &jwks); err != nil {
		return err
	}

	memJWKS = jwks

	// Обновляем kidCount по последнему ключу
	maxKid := 0
	for k := range memJWKS {
		var n int
		fmt.Sscanf(k, "kid-%d", &n)
		if n > maxKid {
			maxKid = n
		}
	}
	kidCount = maxKid

	return nil
}

func GetPrivateKey() rsa.PrivateKey {
	mu.Lock()
	defer mu.Unlock()
	return *memKeys.PrivateKey
}

func GetPublicKey() rsa.PublicKey {
	mu.Lock()
	defer mu.Unlock()
	return *memKeys.PublicKey
}
