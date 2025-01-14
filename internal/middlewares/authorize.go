package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authorizer() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// get role of the context
		role, exists := ctx.Get("role")
		if !exists {
			ctx.Redirect(http.StatusSeeOther, "/public/login")
			return
		}
		// get full path of the request of context
		reqPath := ctx.FullPath()
		if reqPath == "" {
			ctx.Redirect(http.StatusSeeOther, "/public/login")
			return
		}
		// split in array
		splitPath := strings.Split(reqPath, "/")
		if len(splitPath) < 3 {
			ctx.Redirect(http.StatusSeeOther, "/public/login")
			return
		}

		// define the role-path map
		rolePathMapping := map[int64]string{
			1: "student",
			2: "company",
			3: "admin",
			4: "superuser",
			5: "public",
		}
		// check if role and requested path match authority
		if expectedPath, ok := rolePathMapping[role.(int64)]; ok {
			if splitPath[2] != expectedPath {
				ctx.Redirect(http.StatusSeeOther, "/public/login")
				return
			}
		} else {
			ctx.Redirect(http.StatusSeeOther, "/public/login")
			return
		}

		// proceed
		ctx.Next()
	}
}