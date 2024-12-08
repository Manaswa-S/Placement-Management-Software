package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
)

type CompanyHandler struct {
	UserService *services.CompanyService
}

func NewCompanyHandler(userService *services.CompanyService) *CompanyHandler {
	return &CompanyHandler{
		UserService: userService,
	}
}

func (h *CompanyHandler) RegisterRoute(companyRoute *gin.RouterGroup) {
	companyRoute.GET("/getcom", h.CompanyHandlerFunc)
}

func (h *CompanyHandler) CompanyHandlerFunc(c *gin.Context) {

	ctx := c.Request.Context()

	data, err := h.UserService.CompanyS(ctx)
	if err != nil {
		data = "error"
	}
	c.JSON(200, gin.H{
		"handler": "CompanyHandlerFunc",
		"service": data,
	})
}