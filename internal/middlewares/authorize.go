package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
)


func Authorizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("in authorizer...")

		c.Next()
	}
}