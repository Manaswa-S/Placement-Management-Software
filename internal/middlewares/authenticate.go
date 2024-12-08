package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
)


func Authenticator() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("this is the authenticator...")

		c.Next()
	}
}