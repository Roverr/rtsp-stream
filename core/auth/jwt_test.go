package auth

import (
	"fmt"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestJwtAut(t *testing.T) {
	provider := NewJWTProvider("macilaci")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})
	tokenString, err := token.SignedString([]byte("macilaci"))
	assert.Nil(t, err)
	assert.True(t, provider.Validate(tokenString))
	assert.True(t, provider.Validate(fmt.Sprintf("Bearer %s", tokenString)))
}
