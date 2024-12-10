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
	companyRoute.GET("/getcom", h.CompanyHandlerFunc)
}

func (h *CompanyHandler) CompanyHandlerFunc(c *gin.Context) {

	h.CompanyService.CompanyFunc()
	c.JSON(200, gin.H{
		"handler": "CompanyHandlerFunc",
	})
}