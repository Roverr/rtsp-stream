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
	Validate(token string) bool
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
func (jp JWTProvider) Validate(tokenString string) bool {
	ts := strings.Replace(tokenString, "Bearer ", "", -1)
	token, err := jwt.Parse(ts, jp.verify)
	if err != nil {
		logrus.Errorln("Error at token verification ", err)
		return false
	}
	return token.Valid
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
