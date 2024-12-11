package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
	sqlc "go.mod/internal/sqlc/generate"
)

type PublicHandler struct {
	PublicService *services.PublicService
}

func NewPublicHandler(publicService *services.PublicService) *PublicHandler {
	return &PublicHandler{
		PublicService: publicService,
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

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

func (h *PublicHandler) LoginStatic(c *gin.Context) {
	c.File("./template/loginstatic.html")
}

func (h *PublicHandler) SignupStatic(c *gin.Context) {
	c.File("./template/signupstatic.html")
}

func (h *PublicHandler) LoginPost(ctx *gin.Context){

	// parse incoming data
	var loginData services.UserInputData
	err := ctx.Bind(&loginData)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// call the appropriate service
	userRole, JWTTokens, err := h.PublicService.LoginPost(ctx, loginData)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// send jwt as cookies 
	ctx.SetSameSite(http.SameSiteStrictMode)
	ctx.SetCookie("access_token", JWTTokens.JWTAccess, 0, "", "", true, true)
	ctx.SetSameSite(http.SameSiteStrictMode)
	ctx.SetCookie("refresh_token", JWTTokens.JWTRefresh, 0, "", "", true, true)

	// respond with data/template
	// redirect to respective dashboard
	switch userRole {
		case 1 :
			ctx.Redirect(http.StatusSeeOther, "/laa/student/dashboard")
		case 2 :
			ctx.Redirect(http.StatusSeeOther, "/laa/company/dashboard")
		case 3 :
			ctx.Redirect(http.StatusSeeOther, "/laa/admin/dashboard")
		case 4 :
			ctx.Redirect(http.StatusSeeOther, "/laa/superuser/dashboard")
		default :
			ctx.Redirect(http.StatusSeeOther, "/public/signup")
	}
}

func (h *PublicHandler) SignupPost(c *gin.Context){

	// parse incoming data
	var signupData sqlc.SignupUserParams
	err := c.Bind(&signupData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// call appropriate service method
	err = h.PublicService.SignupPost(c, signupData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// respond with data
	// redirect to respective dashboard
	c.JSON(http.StatusOK, gin.H{
		"status": "signup successful!",
	})
}

