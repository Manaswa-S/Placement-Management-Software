package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
)

type StudentHandler struct {
	StudentService *services.StudentService
}

func NewStudentHandler(studentService *services.StudentService) *StudentHandler {
	return &StudentHandler{
		StudentService: studentService,
	}
}

func (h *StudentHandler) RegisterRoute(studentRoute *gin.RouterGroup) {
	studentRoute.GET("/getstu", h.StudentHandlerFunc)
}

func (h *StudentHandler) StudentHandlerFunc(c *gin.Context) {

	h.StudentService.StudentFunc()

	c.JSON(200, gin.H{
		"handler": "StudentHandlerFunc",
	})
}