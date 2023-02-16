package middleware

import (
	"strings"

	"favor-dao-backend/internal/conf"
	"favor-dao-backend/internal/model"
	"favor-dao-backend/pkg/app"
	"favor-dao-backend/pkg/errcode"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func JWT() gin.HandlerFunc {
	// TODO: optimize get user from a simple service that provide fetch a user info interface.
	db := conf.MustGormDB()
	return func(c *gin.Context) {
		var (
			token string
			ecode = errcode.Success
		)
		if s, exist := c.GetQuery("token"); exist {
			token = s
		} else {
			token = c.GetHeader("Authorization")

			if token == "" || !strings.HasPrefix(token, "Bearer ") {
				response := app.NewResponse(c)
				response.ToErrorResponse(errcode.UnauthorizedTokenError)
				c.Abort()
				return
			}

			token = token[7:]
		}
		if token == "" {
			ecode = errcode.InvalidParams
		} else {
			claims, err := app.ParseToken(token)
			if err != nil {
				switch err.(*jwt.ValidationError).Errors {
				case jwt.ValidationErrorExpired:
					ecode = errcode.UnauthorizedTokenTimeout
				default:
					ecode = errcode.UnauthorizedTokenError
				}
			} else {
				c.Set("UID", claims.UID)
				c.Set("USERNAME", claims.Username)

				user := &model.User{
					Model: &model.Model{
						ID: claims.UID,
					},
				}
				user, _ = user.Get(db)
				c.Set("USER", user)

				if (conf.JWTSetting.Issuer + ":" + user.Salt) != claims.Issuer {
					ecode = errcode.UnauthorizedTokenTimeout
				}
			}
		}

		if ecode != errcode.Success {
			response := app.NewResponse(c)
			response.ToErrorResponse(ecode)
			c.Abort()
			return
		}

		c.Next()
	}
}

func JwtLoose() gin.HandlerFunc {
	// TODO: optimize get user from a simple service that provide fetch a user info interface.
	db := conf.MustGormDB()
	return func(c *gin.Context) {
		token, exist := c.GetQuery("token")
		if !exist {
			token = c.GetHeader("Authorization")
			if strings.HasPrefix(token, "Bearer ") {
				token = token[7:]
			} else {
				c.Next()
			}
		}
		if len(token) > 0 {
			if claims, err := app.ParseToken(token); err == nil {
				c.Set("UID", claims.UID)
				c.Set("USERNAME", claims.Username)
				user := &model.User{
					Model: &model.Model{
						ID: claims.UID,
					},
				}
				user, err := user.Get(db)
				if err == nil && (conf.JWTSetting.Issuer+":"+user.Salt) == claims.Issuer {
					c.Set("USER", user)
				}
			}
		}
		c.Next()
	}
}
