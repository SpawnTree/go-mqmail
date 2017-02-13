package main

import (
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestRabbitMQ(t *testing.T) {
	initRabbitMQ()
	assert.Equal(t, RequestQueue, requestQ.Name)
}

func TestSend(t *testing.T) {
	initRabbitMQ()
	msg := "{'name':'Dhanush'}"
	pub := amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(msg),
	}
	err := ch.Publish(ExchangeName, RequestRoutingKey, false, false, pub)
	assert.Nil(t, err)
}

func TestConsume(t *testing.T) {
	initRabbitMQ()
	msgs, err := ch.Consume(
		requestQ.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	assert.Nil(t, err)
	go func() {
		for log := range msgs {
			log.Ack(false)
		}
	}()

}
