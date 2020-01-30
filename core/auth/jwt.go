package auth

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Roverr/rtsp-stream/core/config"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
)

// JWT interface describes how token validation looks like
type JWT interface {
	Validate(token string) (*jwt.Token, *Claim)
}

// Claim describes the claim for the token
type Claim struct {
	Secret string `json:"secret"`
}

// Valid shows if the claim is valid or not
func (c Claim) Valid() error {
	return nil
}

// JWTProvider implements the validate method
type JWTProvider struct {
	secret    []byte
	verifyKey *rsa.PublicKey
}

// Implementation check
var _ JWT = (*JWTProvider)(nil)

// NewJWTProvider returns a new pointer for the created provider
func NewJWTProvider(settings config.Auth) (*JWTProvider, error) {
	switch strings.ToLower(settings.JWTMethod) {
	case "rsa":
		verifyBytes, err := ioutil.ReadFile(settings.JWTPubKeyPath)
		if err != nil {
			return nil, err
		}
		verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
		if err != nil {
			return nil, err
		}
		return &JWTProvider{verifyKey: verifyKey}, nil
	default:
		return &JWTProvider{secret: []byte(settings.JWTSecret)}, nil
	}
}

// Validate is for validating if the given token is authenticated
func (jp JWTProvider) Validate(tokenString string) (*jwt.Token, *Claim) {
	ts := strings.Replace(tokenString, "Bearer ", "", -1)
	if ts == "" {
		logrus.Debug("No token found")
		return nil, nil
	}
	claims := &Claim{}
	token, err := jwt.ParseWithClaims(ts, claims, jp.verify)
	if err != nil {
		logrus.Errorf("Error at token verification: %s | JWTProvider", err)
		return nil, nil
	}
	return token, claims
}

// verify is to check the signing method and return the secret
func (jp JWTProvider) verify(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
		return jp.secret, nil
	}
	if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
		return jp.verifyKey, nil
	}
	return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
}
