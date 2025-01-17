package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mod/internal/dto"
)

func RecordReport(ctx *gin.Context) {
	userid, exists := ctx.Get("ID")
	if !exists {
		ctx.Set("error", "failed to record report : user ID not found in request")
		return
	}
	data := new(dto.Report)
	
	err := ctx.Bind(data)
	if err != nil {
		ctx.Set("error", "failed to record report : cannot bind incoming request")
		return
	}
	data.UserId = userid.(int64)
	data.ReportedAt = time.Now()
	data.IpAddress = ctx.ClientIP()

	jsonLog, _ := json.Marshal(data)
	
	f, err := os.OpenFile("./texts/reports.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		ctx.Set("error", fmt.Sprintf("error opening reports file: %v", err.Error()))
		return
	}

	defer f.Close()

	_, err = f.WriteString(string(jsonLog) + "\n")
	if err != nil {
		ctx.Set("error", fmt.Sprintf("Error writing report: %v", err.Error()))
		return
	}
}