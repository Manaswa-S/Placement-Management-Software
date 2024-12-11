package utils

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	Issuer string
	Subject string
	ExpiresAt int64
	IssuedAt int64
	Role int64
	ID int64	
}


func GenerateJWT(tokenData Token) (string, error) {

	//
	fmt.Println("Generating JWT...")


	// generate a jwt token
	_t_unsigned := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": tokenData.Issuer,
			"sub": tokenData.Subject,
			"exp": tokenData.ExpiresAt,
			"iat": tokenData.IssuedAt,
			"role": tokenData.Role,
			"id": tokenData.ID,
	})

	token, err := _t_unsigned.SignedString([]byte(os.Getenv("SigningKey")))
	if err != nil {
		return "", err
	}

	return token, err
}