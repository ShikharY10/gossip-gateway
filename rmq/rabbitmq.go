package rmq

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/ShikharY10/gbWebSocket/redisAction"
	"github.com/streadway/amqp"
)

type RMQ struct {
	Msgs    <-chan amqp.Delivery
	RedisDB *redisAction.Redis
	ch      *amqp.Channel
	q       *amqp.Queue
}

func (r *RMQ) Init(rmqIP string, username string, password string, name string) {
	var address string = "amqp://" + username + ":" + password + "@" + rmqIP + ":5672/"
	conn, err := amqp.Dial(address)
	if err != nil {
		fmt.Println("[ERROR] : ", err.Error())
	}
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("[ERROR] : ", err.Error())
	}
	r.ch = ch
	q, err := ch.QueueDeclare(
		name,
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println(err.Error())
	}
	r.q = &q
	r.Msgs, err = r.ch.Consume(
		name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println(err.Error())
	}
}

func (r *RMQ) Produce(job []byte) error {
	name := r.getEngineChannel()
	err := r.ch.Publish(
		"",
		name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        job,
		},
	)
	return err
}

func (r *RMQ) getEngineChannel() string {
	names := r.RedisDB.GetEngineName()
	fmt.Println("LEN: ", len(names))
	randomIndex := rand.Intn(len(names))
	pick := names[randomIndex]
	return pick
}
