package controllers

import (
	"gbGATEWAY/epoll"
	"gbGATEWAY/handler"
	"gbGATEWAY/middleware"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Controller struct {
	Epoll      *epoll.EPOLL
	Handler    *handler.Handler
	Middleware *middleware.Middleware
}

func (ctrl *Controller) WebSocketHandler(c *gin.Context) {
	id := c.Value("id").(string)
	err := ctrl.Handler.Database.IsUserRegistered(id)
	if err != nil {
		return
	}
	conn := ctrl.webSocketHandler(c.Writer, c.Request)

	ctrl.Handler.Cache.SetUserConnectNode(id, ctrl.Handler.Env.GateWayName)

	ctrl.Epoll.Lock.Lock()
	ctrl.Epoll.Clients[id] = conn
	ctrl.Epoll.Lock.Unlock()
}

func (ctrl *Controller) webSocketHandler(w http.ResponseWriter, r *http.Request) (conn *websocket.Conn) {
	// Upgrade connection
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}
	if err := ctrl.Epoll.Add(conn); err != nil {
		log.Println("Failed to add connection: ", err)
		conn.Close()
		return nil
	}
	return conn
}
