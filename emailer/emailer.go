package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/rdmyldz/i2t/tesseract"
	"github.com/streadway/amqp"
)

func failOnErr(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

type message struct {
	Name     string
	Email    string
	Filename string
	Image    []byte
}

func decode(data []byte) (*message, error) {
	r := bytes.NewReader(data)
	var msg message
	dec := gob.NewDecoder(r) // Will read from network.
	err := dec.Decode(&msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func getText(data []byte) ([]string, error) {
	handle, err := tesseract.TessBaseAPICreateWithMonitor("tur+eng")
	if err != nil {
		return nil, err
	}

	return handle.ProcessImageMem(data)
}

func main() {
	fmt.Println("rabbit cons_receive")

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnErr(err, "failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnErr(err, "failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare("hello", false, false, false, false, nil)
	failOnErr(err, "failed to declare a queue")

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	failOnErr(err, "failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			go func(d amqp.Delivery) {
				msg, err := decode(d.Body)
				if err != nil {
					log.Fatal(err)
				}

				text, err := getText(msg.Image)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("mail sent to: %s", msg.Email)
				log.Printf("text: %v", text)
			}(d)

		}
	}()

	log.Println(" [*] waiting for messages. to exit press Ctrl+C")

	<-forever
}
