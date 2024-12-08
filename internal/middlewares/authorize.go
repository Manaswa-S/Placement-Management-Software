package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
)


func Authorizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("this is the authorizer...")
		c.Next()
	}
}