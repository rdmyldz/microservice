package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/streadway/amqp"
)

var ErrForm = fmt.Errorf("all form fields must be filled")

type message struct {
	Name  string
	Email string
	Image []byte
}

// Compile templates on start of the application
var templates = template.Must(template.ParseFiles("index.html"))

// Display the named template
func (a *application) display(w http.ResponseWriter, page string, data interface{}) {
	templates.ExecuteTemplate(w, page+".html", data)
}

func (a *application) uploadFile(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		log.Println(err)
		http.Error(w, ErrForm.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	email := r.FormValue("email")
	name := r.FormValue("name")
	filename := handler.Filename
	log.Printf("%s\n%s\n%s", email, name, filename)

	b, err := io.ReadAll(file)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	m := message{
		Name:  name,
		Email: email,
		Image: b,
	}

	pubData, err := encode(m)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = publish(pubData)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "the text will be mailed")
}

func (a *application) handleHome(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		a.display(w, "index", nil)
	case "POST":
		a.uploadFile(w, r)
	}
}

type application struct {
}

func main() {
	app := &application{}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Upload route
	http.HandleFunc("/", app.handleHome)

	log.Printf("listening on %s", port)
	http.ListenAndServe(":"+port, nil)
}

func encode(msg message) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(msg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func publish(data []byte) error {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("hello", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        data,
	})
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}
	return nil
}
