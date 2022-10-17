package rmq

import (
	"errors"
	"fmt"
	"gbGATEWAY/redisAction"
	"log"
	"math/rand"

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
	redisAction.ShowSucces("RabbitMQ Connected", false)
	// color.Green("RMQ connected!")
	// fmt.Println("RMQ connected!")
}

func (r *RMQ) Produce(job []byte) error {
	name, err := r.getEngineChannel()
	if err != nil {
		return err
	}
	err = r.ch.Publish(
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

func (r *RMQ) getEngineChannel() (string, error) {
	names := r.RedisDB.GetEngineName()
	fmt.Println("LEN: ", len(names))
	if len(names) == 0 {
		fmt.Println("[WARNING] : No engine connected!")
		return "", errors.New("no engine connected")
	}
	randomIndex := rand.Intn(len(names))
	pick := names[randomIndex]
	return pick, nil
}
