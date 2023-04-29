package middleware

import (
	"errors"
	"fmt"
	"gbGATEWAY/handler"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type Middleware struct {
	jWT_ACCESS_TOKEN_SECRET_KEY []byte
	Cache                       *handler.CacheHandler
}

func CreateMiddleware(accessToken []byte, cache *handler.CacheHandler) *Middleware {
	return &Middleware{
		jWT_ACCESS_TOKEN_SECRET_KEY: accessToken,
		Cache:                       cache,
	}
}

// varifies JWT access token and the claims the where set while creating the token
func (mw *Middleware) VarifyAccessToken(token string) (claim jwt.MapClaims, err error) {
	newToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("something went wrong")
		}
		return mw.jWT_ACCESS_TOKEN_SECRET_KEY, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := newToken.Claims.(jwt.MapClaims); ok && newToken.Valid {
		return claims, nil
	} else {
		return nil, errors.New("bad token")
	}
}

func (mw *Middleware) APIV3Authorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("New auth request")
		token := c.Query("token")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, "token not found")
			return
		} else {
			claim, err := mw.VarifyAccessToken(token)
			if err != nil {
				if err.Error() == "Token is expired" {
					c.AbortWithStatusJSON(401, err.Error())
				} else {
					c.AbortWithStatusJSON(400, err.Error())
				}
			} else {
				isTokenValid := mw.Cache.IsTokenValid(claim["id"].(string), token, "access")
				if isTokenValid {
					data := map[string]interface{}{
						"id":       claim["id"].(string),
						"username": claim["username"].(string),
						"role":     claim["role"].(string),
					}
					c.Keys = data
					fmt.Println("New auth reques varified")
					c.Next()
				} else {
					c.AbortWithStatus(401)
				}

			}
		}
	}
}
