package middlewares

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mod/internal/config"
	"go.mod/internal/dto"
	"go.mod/internal/utils"
)


func Authenticator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// parse access token string from cookie in the request
		access_token, err := c.Cookie("access_token")
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/public/login")
			c.Abort()
			return
		}
		// parse refresh token string from cookie in the request
		refresh_token, err := c.Cookie("refresh_token")
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/public/login")
			c.Abort()
			return
		}

		// call the parse method to parse the token
		// returns mapped claims or error
		claims, err := utils.ParseJWT(access_token)
		if err != nil {
			// if token expired
			if (errors.Is(err, jwt.ErrTokenExpired)) {
				// parse refresh token to get claims
				mapClaims, err := utils.ParseJWT(refresh_token)
				if err != nil {
					c.Redirect(http.StatusSeeOther, "/public/login")
					c.Abort()
					return
				}
				// generate new access token using refresh token claims
				new_access_token, err := utils.GenerateJWT(dto.Token{
					Issuer: "loginFunc@PMS",
					Subject: "access_token",
					ExpiresAt: time.Now().Add(config.JWTAccessExpiration * time.Second).Unix(),
					IssuedAt: time.Now().Unix(),
					Role: int64(mapClaims["role"].(float64)),
					ID: int64(mapClaims["id"].(float64)),
				})
				if err != nil {
					c.Redirect(http.StatusSeeOther, "/public/login")
					c.Abort()
					return
				}
				// generate new refresh token using refresh token claims
				new_refresh_token, err := utils.GenerateJWT(dto.Token{
					Issuer: "loginFunc@PMS",
					Subject: "refresh_token",
					ExpiresAt: time.Now().Add(config.JWTRefreshExpiration * time.Second).Unix(),
					IssuedAt: time.Now().Unix(),
					Role: int64(mapClaims["role"].(float64)),
					ID: int64(mapClaims["id"].(float64)),
				})
				if err != nil {
					c.Redirect(http.StatusSeeOther, "/public/login")
					c.Abort()
					return
				}
				// set cookies for tokens
				c.SetSameSite(http.SameSiteStrictMode)
				c.SetCookie("access_token", new_access_token, 0, "", "", true, true)
				c.SetSameSite(http.SameSiteStrictMode)
				c.SetCookie("refresh_token", new_refresh_token, 0, "", "", true, true)
				// redirect to the same url to reload and send tokens
				c.Redirect(http.StatusFound, c.Request.URL.String())
			} else {
				c.Redirect(http.StatusSeeOther, "/public/login")
			}	
			// abort to prevent further middlewares from acting
			c.AbortWithStatus(http.StatusFound)
		} else {
			// token is NOT expired
			// set values in context for downstream users
			c.Set("ID", int64(claims["id"].(float64)))
			c.Set("role", int64(claims["role"].(float64)))
			//proceed
			c.Next()
		}
	}
}