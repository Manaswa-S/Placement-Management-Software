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
	studentRoute.GET("/dashboard", h.StudentDashboard)
}

func (h *StudentHandler) StudentDashboard(ctx *gin.Context) {
	ctx.File("./template/dashboard/studentdashboard.html")
}