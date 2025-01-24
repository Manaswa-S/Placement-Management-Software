package utils

import (
	"fmt"
)

func HitProtected(endP string) (error) {
	fmt.Println("hitting", endP, "internally")

	// we will need tokens to access protected endpoints
	
	// 1) directly create tokens here using utils.GenerateJWT
	// this essentially breaks the principle of least priviledge and completely bypasses the user authentication

	// 2) have pre-defined tokens in env files
	// can be used but the tokens by default need to have a short expiry 
	// that means they should be automatically renewed and that needs to have a refreshToken external function
	// so there you go back to point 1 plus some more tangled complexity

	// 3) first get token by logging in or something like that and use that token
	// have admin credentials in env, use them to call the log in endpoint which returns tokens
	// use these tokens for future use and atlast logout or do something to avoid spamming or cluttering user instance




	// accTk := dto.Token{
	// 	Issuer: "internal_caller",
	// 	Subject: "access_token",
	// 	ExpiresAt: time.Now().Add(config.JWTAccessExpiration * time.Minute).Unix(),
	// 	IssuedAt: time.Now().Unix(),
	// 	Role: 3,
	// 	ID: ,
	// }


	// GenerateJWT()








	return nil
}