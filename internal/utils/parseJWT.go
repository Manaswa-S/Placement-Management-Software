package utils

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)


func ParseJWT(tokenString string) (jwt.MapClaims, error) {
		// parse, validate and verify token string signature
		token, err := jwt.Parse(tokenString, 
				func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(os.Getenv("SigningKey")), nil
		})
		if err != nil {
			return nil, err
		}

		// jwt auto checks for expiry
		// TODO: add expiry check for extra care


		// token is valid
		// map tokens to respective keys
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			return nil, errors.New("invalid token string")
		}

		return claims, err
}
