package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
)

type OpenHandler struct {
	OpenService *services.OpenService
}

func NewOpenHandler(openService *services.OpenService) *OpenHandler {
	return &OpenHandler{
		OpenService: openService,
	}
}

func (h *OpenHandler) RegisterRoute(openRoute *gin.RouterGroup) {
	
}


