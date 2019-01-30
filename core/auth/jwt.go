package auth

import (
	"fmt"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
)

// JWT interface describes how token validation looks like
type JWT interface {
	Validate(token string) bool
}

// JWTProvider implements the validate method
type JWTProvider struct {
	secret []byte
}

// Implementation check
var _ JWT = (*JWTProvider)(nil)

// NewJWTProvider returns a new pointer for the created provider
func NewJWTProvider(secret string) *JWTProvider {
	return &JWTProvider{[]byte(secret)}
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
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}
	return jp.secret, nil
}
