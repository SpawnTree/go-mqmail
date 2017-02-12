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

func TestSendRabbitMQ(t *testing.T) {
	initRabbitMQ()
	msg := "{'name':'Dhanush'}"
	pub := amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(msg),
	}
	err := ch.Publish(ExchangeName, "emails", false, false, pub)
	assert.Nil(t, err)
}
