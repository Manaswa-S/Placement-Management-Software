package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
)

type PublicHandler struct {
	UserService *services.PublicService
}

func NewPublicHandler(userService *services.PublicService) *PublicHandler {
	return &PublicHandler{
		UserService: userService,
	}
}

func (h *PublicHandler) RegisterRoute(publicRoute *gin.RouterGroup) {
	// get the login static page, direct
	publicRoute.GET("/login", h.LoginStatic)
	// get the login static page, direct
	publicRoute.GET("/signup", h.SignupStatic)
	// post the data from login page, indirect
	publicRoute.POST("/postlogindata", h.LoginPost)
	// post the data from signup page, indirect
	publicRoute.POST("/postsignupdata", h.SignupPost)
}

func (h *PublicHandler) LoginStatic(c *gin.Context) {
	c.File("./template/loginstatic.html")
}

func (h *PublicHandler) SignupStatic(c *gin.Context) {
	c.File("./template/signupstatic.html")
}

func (h *PublicHandler) LoginPost(c *gin.Context){

	var loginData struct {
		Email string
		Password string
	}

	err := c.Bind(&loginData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": loginData,
	})
}


type SignupData struct {
	Email string
	Password string
	Role string
}


func (h *PublicHandler) SignupPost(c *gin.Context){

	var signupData SignupData

	err := c.Bind(&signupData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	ctx := c.Request.Context()

	userData, err := h.UserService.PublicS(ctx, &signupData)

	c.JSON(http.StatusOK, gin.H{
		"data": signupData,
	})
}

