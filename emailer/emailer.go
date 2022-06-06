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
	dec := gob.NewDecoder(r)
	err := dec.Decode(&msg)
	if err != nil {
		return nil, fmt.Errorf("decoding err: %w", err)
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

type Result struct {
	name     string
	filename string
	email    string
	content  []string
	err      error
}

func sendEmail(name, filename, email string, content []string) {
	log.Printf("email is sending to %s", email)
	log.Printf("\nDear %s,\nthe file %s's content is attached\n\n%s\n",
		name, filename, content)
}

func main() {
	fmt.Println("rabbit cons_receive")

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnErr(err, "failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnErr(err, "failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare("img_text", false, false, false, false, nil)
	failOnErr(err, "failed to declare a queue")

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	failOnErr(err, "failed to register a consumer")

	forever := make(chan bool)

	result := make(chan *Result)
	go func() {
		defer close(result)
		for res := range result {
			if res.err != nil {
				log.Printf("something happened. Email won't be sent: %v", res.err)
				continue
			}
			sendEmail(res.name, res.filename, res.email, res.content)
		}
	}()

	go func() {
		for d := range msgs {
			go func(d amqp.Delivery) {
				var res Result
				msg, err := decode(d.Body)
				if err != nil {
					res.err = err
					result <- &res
					return
				}
				res.name = msg.Name
				res.filename = msg.Filename
				res.email = msg.Email

				text, err := getText(msg.Image)
				if err != nil {
					res.err = err
					result <- &res
					return
				}
				res.content = text
				result <- &res
			}(d)

		}
	}()

	log.Println(" [*] waiting for messages. to exit press Ctrl+C")

	<-forever
}
