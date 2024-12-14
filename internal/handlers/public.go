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

	// get the reset pass static page, direct
	publicRoute.GET("/resetpassgetemail", h.ResetPassGetEmail)
	// get the reset pass static page, direct
	publicRoute.GET("/resetpassgetpass", h.ResetPassGetPass)
	// get the login static page, direct
	publicRoute.POST("/resetpasspostemail", h.ResetPassPostEmail)
	// get the login static page, direct
	publicRoute.POST("/resetpasspostpass", h.ResetPassPostPass)

	// post the data from login page, indirect
	publicRoute.POST("/postlogindata", h.LoginPost)
	// post the data from signup page, indirect
	publicRoute.POST("/postsignupdata", h.SignupPost)
	// send email for email id confirmation
	publicRoute.GET("/sendconfirmemail", h.SendConfirmationEmail)
	// confirm email id, validate
	publicRoute.GET("/confirmsignup", h.ConfirmSignup)
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

func (h *PublicHandler) LoginStatic(ctx *gin.Context) {
	ctx.File("./template/loginstatic.html")
}

func (h *PublicHandler) SignupStatic(ctx *gin.Context) {
	ctx.File("./template/signupstatic.html")
}

func (h *PublicHandler) ResetPassGetEmail(ctx *gin.Context) {
	ctx.File("./template/reset/passresetgetemail.html")
}

func (h *PublicHandler) ResetPassPostEmail(ctx *gin.Context) {
	
	// get email of request
	var email struct {
		Email string
	} // bind
	err := ctx.Bind(&email)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// call service
	err = h.PublicService.SendResetPassEmail(ctx, email.Email)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// respond
	ctx.JSON(http.StatusOK, gin.H{
		"status": "pass reset email sent successfully",
	})
}

func (h *PublicHandler) ResetPassGetPass(ctx *gin.Context) {

	// call service
	body, err := h.PublicService.GetPass(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	// respond
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", body.Bytes())
}

func (h *PublicHandler) ResetPassPostPass(ctx *gin.Context) {
	
	var data services.ResetPass

	// data.Token = ctx.Query("token")
	// bind data
	err := ctx.Bind(&data)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// call service
	err = h.PublicService.ResetPass(ctx, data)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// respond
	ctx.JSON(http.StatusOK, gin.H{
		"status": "password reset successfully. please proceed to log in",
	})
}

func (h *PublicHandler) SignupPost(ctx *gin.Context){

	// parse incoming data
	var signupData sqlc.SignupUserParams
	err := ctx.Bind(&signupData)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// call appropriate service method
	err = h.PublicService.SignupPost(ctx, signupData)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// respond with data
	// redirect to respective dashboard
	ctx.JSON(http.StatusOK, gin.H{
		"status": "signup successful! confirm email and then log in.",
	})
}

func (h *PublicHandler) SendConfirmationEmail(ctx *gin.Context){

	// get email of request
	email := ctx.Query("email")
	if email == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "email is required",
		})
		return
	}
	// call service
	err := h.PublicService.SendConfirmEmail(ctx, email)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "not able to  send confirmation email. try again",
		})
		return
	}
	// respond
	ctx.JSON(http.StatusOK, gin.H{
		"status": "confirmation email sent",
	})
}

func(h *PublicHandler) ConfirmSignup(ctx *gin.Context) {
	
	// get token from query 
	confirmToken := ctx.Query("token")
	if confirmToken == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Invalid token. Try again!",
		})	
		return
	}

	// call service
	err := h.PublicService.ConfirmEmail(ctx, confirmToken)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})	
		return
	}
	
	// respond
	ctx.JSON(200, gin.H{
		"status": "email confirmed. proceed to login.",
	})
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