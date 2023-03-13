package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type Middleware struct {
	SecretKey []byte
}

func varifyJWT(token string, secretekey []byte) (jwt.MapClaims, error) {
	newToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("something went wrong")
		}
		return secretekey, nil
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
		fmt.Println("New Request=== " + c.Request.URL.Path + " ===")
		token := c.Query("token")
		fmt.Println("Token: ", token)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, "token not found")
			return
		} else {
			claim, err := varifyJWT(token, mw.SecretKey)
			if err != nil {
				c.Header("error", err.Error())
				c.AbortWithStatusJSON(400, "invalid JWT Token, "+err.Error())
				return
			} else {
				data := map[string]interface{}{
					"uuid":     claim["uuid"].(string),
					"username": claim["username"].(string),
					"role":     claim["role"].(string),
				}
				c.Keys = data
				fmt.Println("=== Request Varified ===")
				c.Next()
			}
		}
	}
}
