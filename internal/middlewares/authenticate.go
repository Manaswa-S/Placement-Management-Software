package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mod/internal/utils"
)


func Authenticator() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("in authenticator")
		// parse token string from cookie in the request
		access_token, err := c.Cookie("access_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid access_token. log in again",
			})
		}

		// call the parse method to parse the token
		// returns mapped claims or error
		claims, err := utils.ParseJWT(access_token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid access_token. log in again",
			})
		}

		// if valid token
		// set values in context for downstream users
		c.Set("ID", int64(claims["id"].(float64)))
		c.Set("role", int64(claims["role"].(float64)))

		//proceed
		c.Next()
	}
}