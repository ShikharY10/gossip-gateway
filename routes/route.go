package routes

import (
	"gbGATEWAY/controllers"

	"github.com/gin-gonic/gin"
)

func WebsocketRoute(router *gin.Engine, controller controllers.Controller) {
	secure := router.Group("/")
	secure.Use(controller.Middleware.APIV3Authorization())
	secure.GET("/connect", controller.WebSocketHandler)
}
