package controllers

import (
	"encoding/json"
	"gbGATEWAY/epoll"
	"gbGATEWAY/handler"
	"gbGATEWAY/middleware"
	"gbGATEWAY/schema"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
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
		c.AbortWithStatusJSON(404, gin.H{
			"statusstring": err.Error(),
		})
		return
	}
	conn, err := ctrl.webSocketHandler(c.Writer, c.Request)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"statusstring": err.Error(),
		})
		return
	}

	err = ctrl.Handler.Cache.SetUserConnectNode(id, ctrl.Handler.Env.GateWayName)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"statusstring": err.Error(),
		})
		return
	}

	ctrl.Epoll.Lock.Lock()
	ctrl.Epoll.Clients[id] = conn
	ctrl.Epoll.Lock.Unlock()

	type PendingPacket struct {
		UserId string `json:"userId"`
	}

	pendingPacket := PendingPacket{
		UserId: id,
	}
	pendingPacketByte, err := json.Marshal(&pendingPacket)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"statusstring": err.Error(),
		})
		return
	}

	payload := schema.Payload{
		Data: pendingPacketByte,
		Type: "002",
	}
	payloadBytes, err := proto.Marshal(&payload)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"statusstring": err.Error(),
		})
		return
	}

	engine, err := ctrl.Handler.Cache.GetRandomEngineName()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"statusstring": err.Error(),
		})
		return
	}

	err = ctrl.Handler.Queue.Produce(engine, payloadBytes)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"statusstring": err.Error(),
		})
		return
	}
}

func (ctrl *Controller) webSocketHandler(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	// Upgrade connection
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	if err := ctrl.Epoll.Add(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}
