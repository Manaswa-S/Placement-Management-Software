package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response middleware that handles all response formatting
func ResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute the handler
		c.Next()

		// Check if there's any error set in the context
		if len(c.Errors) > 0 {
			// Get the error
			err := c.Errors.Last().Err
			// Send a custom error response if there is one
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// If no error, check if there's a response to send (data, not error)
		if c.Writer.Status() == http.StatusOK {
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
			})
		}
	}
}
