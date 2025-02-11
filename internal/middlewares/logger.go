package middlewares

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mod/internal/dto"
)




func Logger(errorsChan chan *dto.ErrorData) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		startTime := time.Now()

		ctx.Next()

		logData := dto.LoggerData {
			StartTime: startTime,
			ClientIP: ctx.ClientIP(),
			Method: ctx.Request.Method,
			Path: ctx.Request.URL.Path,
			StatusCode: ctx.Writer.Status(),
			InternalError: ctx.Errors.String(),
			Latency: time.Duration(time.Since(startTime).Microseconds()),
		}

		errorCheck(ctx, errorsChan, &logData)

		jsonLog, _ := json.Marshal(logData)

		// TODO: dont onen and close the file every time 
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


func errorCheck(ctx *gin.Context, errorsChan chan *dto.ErrorData, logData *dto.LoggerData) {

	errData := new(dto.ErrorData)
	var erred bool

	if debugErr, exists := ctx.Get("debug"); exists {
		errData.Debug = fmt.Sprintf("%s", debugErr)
		erred = true
	} 
	if infoErr, exists := ctx.Get("info"); exists {
		errData.Info = fmt.Sprintf("%s", infoErr)
		erred = true
	}
	if warnErr, exists := ctx.Get("warn"); exists {
		errData.Warn = fmt.Sprintf("%s", warnErr)
		erred = true
	}
	if errorErr, exists := ctx.Get("error"); exists {
		errData.Error = fmt.Sprintf("%s", errorErr)
		erred = true
	}
	if criticalErr, exists := ctx.Get("critical"); exists {
		errData.Critical = fmt.Sprintf("%s", criticalErr)
		erred = true
	}
	if fatalErr, exists := ctx.Get("fatal"); exists {
		errData.Fatal = fmt.Sprintf("%s", fatalErr)
		erred = true
	}
	if erred {
		errData.LogData = logData
		errorsChan <- errData
		// TODO: have another logging file for explicit errors
	}
} 

