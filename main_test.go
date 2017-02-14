package main

import (
	"testing"

	"encoding/json"

	"os"

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

func TestPublishAndConsumeMail(t *testing.T) {
	initRabbitMQ()
	requestBody := NewRequest([]string{"dhanushgopi@yahoo.com"}, nil, nil, "e1@mailinator.com", "Hello World!", "Hello", true)
	requestBodyJSON, e := json.Marshal(requestBody)
	assert.Nil(t, e)
	pub := amqp.Publishing{
		ContentType: "application/json",
		Body:        requestBodyJSON,
	}
	err := ch.Publish(ExchangeName, RequestRoutingKey, false, false, pub)
	assert.Nil(t, err)

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

	done := make(chan bool)
	os.Setenv("GO_ENV", "test")
	go func() {
		for log := range msgs {
			req := &Request{}
			err := json.Unmarshal(log.Body, req)
			assert.Nil(t, err)
			assert.Equal(t, req.Subject, "Hello World!")
			ok, err := req.SendEmail()
			assert.True(t, ok)
			assert.Nil(t, err)
			log.Ack(true)
		}
		done <- true
	}()
	isDone := <-done
	assert.True(t, isDone)
}

func BenchmarkPubAndConsumeMail(b *testing.B) {
	initRabbitMQ()
	os.Setenv("GO_ENV", "test")
	for n := 0; n < b.N; n++ {
		requestBody := NewRequest([]string{"dhanushgopi@yahoo.com"}, nil, nil, "e1@mailinator.com", "Hello World!", "Hello", true)
		requestBodyJSON, _ := json.Marshal(requestBody)
		pub := amqp.Publishing{
			ContentType: "application/json",
			Body:        requestBodyJSON,
		}
		ch.Publish(ExchangeName, RequestRoutingKey, false, false, pub)

		msgs, _ := ch.Consume(
			requestQ.Name,
			"",
			true,
			false,
			false,
			false,
			nil,
		)
		go func() {
			for log := range msgs {
				req := &Request{}
				json.Unmarshal(log.Body, req)
				req.SendEmail()
				log.Ack(true)
			}
		}()
	}

}
