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
			ctx.Abort()
			return
		}
		// get full path of the request of context
		reqPath := ctx.FullPath()
		if reqPath == "" {
			ctx.Redirect(http.StatusSeeOther, "/public/login")
			ctx.Abort()
			return
		}
		// split in array
		splitPath := strings.Split(reqPath, "/")
		if len(splitPath) < 3 {
			ctx.Redirect(http.StatusSeeOther, "/public/login")
			ctx.Abort()
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

		openPathMapping := map[int64]string{
			1: "open",
			2: "open",
			3: "open",
			4: "open",
		}

		// check if role and requested path match authority
		// TODO: fix this, too chaotic, simplify yet secure
		requestedPath := splitPath[2]
		if expectedPath, ok := rolePathMapping[role.(int64)]; ok {
			if requestedPath != expectedPath {
				if openPath, ok := openPathMapping[role.(int64)]; ok {
					if requestedPath != openPath {
						ctx.Redirect(http.StatusSeeOther, "/public/login")
						ctx.Abort()
						return
					}
				}
			}
		} else {
			ctx.Redirect(http.StatusSeeOther, "/public/login")
			ctx.Abort()
			return
		}

		// proceed
		ctx.Next()
	}
}