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
		// You can customize the response data formatting as per your requirements
		if c.Writer.Status() == http.StatusOK {
			// Example of sending a success response with the data
			// If the handler has set some data to be sent, you could use:
			// c.JSON(http.StatusOK, c.Get("data"))
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
			})
		}
	}
}
