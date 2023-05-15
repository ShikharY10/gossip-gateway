package handler

import (
	"gbGATEWAY/admin"
	"gbGATEWAY/config"
	"gbGATEWAY/schema"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"
)

type QueueHandler struct {
	Queue   config.Queue
	Clients map[string]*websocket.Conn
	Logger  *admin.Logger
}

func (queue *QueueHandler) Produce(nodeName string, data []byte) error {

	err := queue.Queue.Channel.Publish(
		"",
		nodeName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		},
	)
	return err
}

func (queue *QueueHandler) Consume() {
	var deliveryPacket schema.DeliveryPacket

	for jobs := range queue.Queue.Jobs {
		err := proto.Unmarshal(jobs.Body, &deliveryPacket)
		if err != nil {
			// ToDo: log this error
			continue
		}

		conn := queue.Clients[deliveryPacket.TargetId]

		if conn != nil {
			err = conn.WriteMessage(2, deliveryPacket.Payload)
			if err != nil {
				// ToDo: log this error
				continue
			}
		}
	}
}
