package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
)

type SuperUserHandler struct {
	SuperService *services.SuperService
}

func NewSuperUserHandler(superService *services.SuperService) *SuperUserHandler {
	return &SuperUserHandler{
		SuperService: superService,
	}
}

func (h *SuperUserHandler) RegisterRoute(superuserRoute *gin.RouterGroup) {
	superuserRoute.GET("/dashboard", h.SuperDashboard)
}

func (h *SuperUserHandler) SuperDashboard(ctx *gin.Context) {
	ctx.File("./template/dashboard/superdashboard.html")	
}