package utils

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
	"go.mod/internal/dto"
)

func GenerateJWT(tokenData dto.Token) (string, error) {
	// generate a jwt token
	_t_unsigned := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": tokenData.Issuer,
			"sub": tokenData.Subject,
			"exp": tokenData.ExpiresAt,
			"iat": tokenData.IssuedAt,
			"role": tokenData.Role,
			"id": tokenData.ID,
			"email": tokenData.Email,
	})

	token, err := _t_unsigned.SignedString([]byte(os.Getenv("SigningKey")))
	if err != nil {
		return "", err
	}

	return token, err
}