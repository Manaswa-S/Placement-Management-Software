package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
)

type StudentHandler struct {
	UserService *services.StudentService
}

func NewStudentHandler(userService *services.StudentService) *StudentHandler {
	return &StudentHandler{
		UserService: userService,
	}
}

func (h *StudentHandler) RegisterRoute(studentRoute *gin.RouterGroup) {
	studentRoute.GET("/getstu", h.StudentHandlerFunc)
}

func (h *StudentHandler) StudentHandlerFunc(c *gin.Context) {

	ctx := c.Request.Context()

	data, err := h.UserService.StudentS(ctx)
	if err != nil {
		data = "error"
	}
	c.JSON(200, gin.H{
		"handler": "StudentHandlerFunc",
		"service": data,
	})
}