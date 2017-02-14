package main

import (
	"log"

	"github.com/streadway/amqp"
)

const (
	//RabbitMQUrl stores the url
	RabbitMQUrl        string = "amqp://guest:guest@localhost:5672/"
	RequestQueue       string = "email.request"
	ResponseQueue      string = "email.response"
	ExchangeName       string = "MQMail"
	RequestRoutingKey  string = "requestRoutingKey"
	ResponseRoutingKey string = "responseRoutingKey"
)

var (
	conn      *amqp.Connection
	ch        *amqp.Channel
	requestQ  amqp.Queue
	responseQ amqp.Queue
)

func initRabbitMQ() {
	conn, err := amqp.Dial(RabbitMQUrl)
	if err != nil {
		log.Printf("Error in connecting\n")
		panic(err)
	}

	ch, err = conn.Channel()
	if err != nil {
		log.Printf("Cannot create channel \n")
		panic(err)
	}
	err = ch.ExchangeDeclare(
		ExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed Exchange declaration \n")
		panic(err)
	}
	requestQ, err = ch.QueueDeclare(
		RequestQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed Request Queue declaration \n")
		panic(err)
	}
	responseQ, err = ch.QueueDeclare(
		ResponseQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed Response Queue declaration \n")
		panic(err)
	}

	err = ch.QueueBind(
		requestQ.Name,
		RequestRoutingKey,
		ExchangeName,
		false,
		nil)
	if err != nil {
		log.Printf("Failed Request Queue Binding \n")
		panic(err)
	}
	err = ch.QueueBind(
		responseQ.Name,
		ResponseRoutingKey,
		ExchangeName,
		false,
		nil)
	if err != nil {
		log.Printf("Failed Response Queue Binding \n")
		panic(err)
	}

}

func main() {
	initRabbitMQ()
	defer conn.Close()
	log.Println("Connected with RabbitMQ")
	log.Println("Waiting for the messages to arrive.")
	for true {

	}
}
