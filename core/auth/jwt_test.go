package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"

	"github.com/Roverr/rtsp-stream/core/config"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestJWTAuthWithSecret(t *testing.T) {
	spec := config.InitConfig()
	provider, err := NewJWTProvider(spec.Auth)
	assert.Nil(t, err)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})
	tokenString, err := token.SignedString([]byte(spec.Auth.JWTSecret))
	assert.Nil(t, err)
	validated, _ := provider.Validate(tokenString)
	assert.NotNil(t, validated)
	validated, _ = provider.Validate(fmt.Sprintf("Bearer %s", tokenString))
	assert.NotNil(t, validated)
}

func TestJWTAuthWithRSA(t *testing.T) {
	reader := rand.Reader
	bitSize := 2048
	key, err := rsa.GenerateKey(reader, bitSize)
	assert.Nil(t, err)
	publicKey := key.PublicKey
	provider := JWTProvider{verifyKey: &publicKey}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{})
	tokenString, err := token.SignedString(key)
	assert.Nil(t, err)
	validated, _ := provider.Validate(tokenString)
	assert.NotNil(t, validated)
}
