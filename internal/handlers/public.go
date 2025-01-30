package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	errs "go.mod/internal/const"
	"go.mod/internal/services"
	sqlc "go.mod/internal/sqlc/generate"
	"go.mod/internal/utils"
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
	publicRoute.GET("/login", h.LoginStatic) //
	// get the login static page, direct
	publicRoute.GET("/signup", h.SignupStatic) //

	// get the reset pass static page, direct
	publicRoute.GET("/resetpassgetemail", h.ResetPassGetEmail) //
	// get the reset pass static page, direct
	publicRoute.GET("/resetpassgetpass", h.ResetPassGetPass) //
	// get the login static page, direct
	publicRoute.POST("/resetpasspostemail", h.ResetPassPostEmail) //
	// get the login static page, direct
	publicRoute.POST("/resetpasspostpass", h.ResetPassPostPass) //

	// post the data from login page, indirect
	publicRoute.POST("/postlogindata", h.LoginPost) //
	// post the data from signup page, indirect
	publicRoute.POST("/postsignupdata", h.SignupPost) //
	// send email for email id confirmation
	publicRoute.GET("/sendconfirmemail", h.SendConfirmationEmail) //
	// confirm email id, validate
	publicRoute.GET("/confirmsignup", h.ConfirmSignup) //
	// post the data from extra info page, indirect
	publicRoute.POST("/extrainfopost", h.ExtraInfoPost) //
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

func (h *PublicHandler) LoginStatic(ctx *gin.Context) {
	ctx.File("./template/public/loginstatic.html")
}

func (h *PublicHandler) SignupStatic(ctx *gin.Context) {
	ctx.File("./template/public/signupstatic.html")
}

func (h *PublicHandler) ResetPassGetEmail(ctx *gin.Context) {
	ctx.File("./template/public/passresetgetemail.html")
}

func (h *PublicHandler) ResetPassPostEmail(ctx *gin.Context) {
	// get email of request
	var email struct {
		Email string
	} // bind
	err := ctx.Bind(&email)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	// call service
	err = h.PublicService.SendResetPassEmail(ctx, email.Email)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	// respond
	ctx.Status(http.StatusOK)
}

func (h *PublicHandler) ResetPassGetPass(ctx *gin.Context) {

	// call service
	body, err := h.PublicService.GetPass(ctx)
	if err != nil {
		ctx.Redirect(http.StatusSeeOther, "/public/resetpassgetemail")
		return
	}

	// respond
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", body.Bytes())
}

func (h *PublicHandler) ResetPassPostPass(ctx *gin.Context) {
	var data services.ResetPass

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
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Message": err.Error(),
		})
		return
	}

	// call appropriate service method
	errf := h.PublicService.SignupPost(ctx, signupData)
	if errf != nil {
		if (errf.Type != errs.Internal) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"Type": errf.Type,
				"Message": errf.Message,		
			})
		}
		return
	}

	// respond with data
	// redirect to respective dashboard
	ctx.JSON(http.StatusOK, gin.H{
		"Status": "Signup Initiated! Check email for further instructions.",
	})
}

const (
	RoleStudent = 1
	RoleCompany = 2	
)

func (h *PublicHandler) ExtraInfoPost(ctx *gin.Context) {

	token := ctx.Request.FormValue("Token")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"errors": "empty token received",
		})
		return
	}
	claims, err := utils.ParseJWT(token)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"errors": err.Error(),
		})
	}

	errf := new(errs.Error)
	
	role, ok := claims["role"]
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Type": errs.InvalidState,
			"Message": "Invalid role provided.",
		})
		return
	}
	roleFloat, ok := role.(float64)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Type": errs.InvalidState,
			"Message": "Invalid role provided.",
		})
		return
	}
	
	roleInt := int64(roleFloat) 

	switch roleInt {
	case RoleStudent:
		_, errf = h.PublicService.ExtraInfoPostStudent(ctx, claims)
	case RoleCompany:
		_, errf = h.PublicService.ExtraInfoPostCompany(ctx, claims)
	default:
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"Type": errs.Unauthorized,
			"Message": "invalid role",
		})
		return
	}

	if errf != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Type": errf.Type,
			"Message": errf.Message,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"Status": "Sign up complete. Proceed with further instructions as given in the email.",
	})
}

func (h *PublicHandler) SendConfirmationEmail(ctx *gin.Context){

	// get email of request
	email := ctx.Query("email")
	if email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Message": "Email not found.",
		})
		return
	}
	// call service
	err := h.PublicService.SendConfirmEmail(ctx, email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Message": "Failed to send confirmation email. Try again." + err.Error(),
		})
		return
	}
	// respond
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
		<script>
			alert("Sent email successfully!");
		</script>
	`))
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
	body, err := h.PublicService.ConfirmEmail(ctx, confirmToken)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})	
		return
	}
	// respond
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", body.Bytes())
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
	userRole, JWTTokens, errf := h.PublicService.LoginPost(ctx, loginData)
	if errf != nil {
		if (errf.Type != errs.Internal) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"Type": errf.Type,
				"Message": errf.Message,
			})
		}
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