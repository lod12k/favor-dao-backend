package middleware

import (
	"encoding/json"
	"fmt"

	"favor-dao-backend/internal/conf"
	"favor-dao-backend/internal/core"
	"favor-dao-backend/internal/service"
	"favor-dao-backend/pkg/app"
	"favor-dao-backend/pkg/errcode"
	"github.com/gin-gonic/gin"
)

func Session() gin.HandlerFunc {
	redis := conf.Redis
	return func(c *gin.Context) {
		var (
			token string
			ecode = errcode.Success
		)
		token = c.GetHeader("X-Session-Token")
		if token != "" {
			raw, err := redis.Get(c, fmt.Sprintf("token_%s", token)).Bytes()
			if err == nil {
				var session service.Session
				err = json.Unmarshal(raw, &session)
				if err != nil {
					ecode = errcode.UnauthorizedTokenError
				} else {
					user, err := service.GetUserByAddress(session.WalletAddr)
					if err != nil {
						ecode = errcode.UnauthorizedAuthNotExist
					} else {
						c.Set("USER", user)
						c.Set("address", user.Address)
					}
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

func Login() gin.HandlerFunc {
	redis := conf.Redis
	return func(c *gin.Context) {
		var (
			token string
			ecode = errcode.Success
		)
		token = c.GetHeader("X-Session-Token")

		if token == "" {
			response := app.NewResponse(c)
			response.ToErrorResponse(errcode.UnauthorizedTokenError)
			c.Abort()
			return
		}
		key := fmt.Sprintf("token_%s", token)
		raw, err := redis.Get(c, key).Bytes()
		if err == nil {
			var session service.Session
			err = json.Unmarshal(raw, &session)
			if err != nil {
				ecode = errcode.UnauthorizedTokenError
			} else {
				user, err := service.GetUserByAddress(session.WalletAddr)
				if err != nil {
					ecode = errcode.UnauthorizedAuthNotExist
				} else {
					c.Set("USER", user)
					c.Set("address", user.Address)
				}
			}
		} else {
			ecode = errcode.UnauthorizedTokenTimeout
		}

		if ecode != errcode.Success {
			response := app.NewResponse(c)
			response.ToErrorResponse(ecode)
			c.Abort()
			return
		}
		redis.Expire(c, key, core.TokenExpiration)
		c.Next()
	}
}
