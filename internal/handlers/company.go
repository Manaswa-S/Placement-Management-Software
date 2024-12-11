package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
)

type CompanyHandler struct {
	CompanyService *services.CompanyService
}

func NewCompanyHandler(companyService *services.CompanyService) *CompanyHandler {
	return &CompanyHandler{
		CompanyService: companyService,
	}
}

func (h *CompanyHandler) RegisterRoute(companyRoute *gin.RouterGroup) {
	companyRoute.GET("/dashboard", h.CompanyDashboard)
}

func (h *CompanyHandler) CompanyDashboard(ctx *gin.Context) {
	ctx.File("./template/dashboard/companydashboard.html")
}