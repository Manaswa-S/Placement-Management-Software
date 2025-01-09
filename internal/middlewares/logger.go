package middlewares

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type LoggerFormat struct {
	StartTime time.Time
	ClientIP string
	Method string
	Path string
	StatusCode int
	ErrorMsg string
	Latency time.Duration
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {

		before := LoggerFormat {
			StartTime: time.Now(),
			ClientIP: c.ClientIP(),
			Method: c.Request.Method,
			Path: c.Request.URL.Path,
		}

		c.Next()

		statusCode := c.Writer.Status()
		after := LoggerFormat {
			StatusCode: statusCode,
			ErrorMsg: c.Errors.String(),
			Latency: time.Duration(time.Since(before.StartTime).Microseconds()),
		}

		var explicit_err any
		experr, exists := c.Get("error")
		if exists {
			explicit_err = experr
		} else {
			explicit_err = "nil"
		}

		logData := map[string]interface{}{
			"client_ip": before.ClientIP,
			"method":    before.Method,
			"path":      before.Path,
			"status":    after.StatusCode,
			"internal_error":     after.ErrorMsg,
			"explicit_error":     explicit_err.(string),
			"latency":   after.Latency.String(),
		}
		jsonLog, _ := json.Marshal(logData)
		
		f, err := os.OpenFile("./texts/logger.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("error opening logger file: ", err.Error())
			return
		}

		defer f.Close()

		_, err = f.WriteString(string(jsonLog) + "\n")
		if err != nil {
			fmt.Println("Error writing log: ", err.Error())
			return
		}
	}
}