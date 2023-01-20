package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
)

type BuyFund struct {
	ID     string `json:"investor"`
	Fund   string `json:"fund"`
	Amount int    `json:"amount"`
}

type Sell struct {
	ID   string  `json:"investor"`
	Fund string  `json:"fund"`
	Qty  float64 `json:"qty"`
}

type Order struct {
	ID       string  `json:"order_id"`
	Investor string  `json:"investor"`
	Fund     string  `json:"fund"`
	Qty      float64 `json:"qty"`
}

type InfoMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Config struct {
	RabbitUser string `envconfig:"RABBITMQ_USER"`
	RabbitPwd  string `envconfig:"RABBITMQ_PWD"`
	RabbitHost string `envconfig:"RABBITMQ_HOST"`
	RabbitPort string `envconfig:"RABBITMQ_PORT"`
	BuyQueue   string `envconfig:"BUY_QUEUE"`
	SellQueue  string `envconfig:"SELL_QUEUE"`
	OrderQueue string `envconfig:"ORDER_QUEUE"`
}

var RabbitConnect = ""
var SellQueue = ""
var BuyQueue = ""
var OrderQueue = ""

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	var config Config
	config.RabbitUser = os.Getenv("RABBITMQ_USER")
	config.RabbitPwd = os.Getenv("RABBITMQ_PWD")
	config.RabbitHost = os.Getenv("RABBITMQ_HOST")
	config.RabbitPort = os.Getenv("RABBITMQ_PORT")
	config.BuyQueue = os.Getenv("BUY_QUEUE")
	config.SellQueue = os.Getenv("SELL_QUEUE")
	config.OrderQueue = os.Getenv("ORDER_QUEUE")

	RabbitConnect = "amqp://" + config.RabbitUser + ":" + config.RabbitPwd + "@" + config.RabbitHost + ":" + config.RabbitPort + "/"
	BuyQueue = config.BuyQueue
	SellQueue = config.SellQueue
	OrderQueue = config.OrderQueue

	fmt.Print(RabbitConnect)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/new/invest", NewInvest).Methods("POST")
	router.HandleFunc("/new/sell", NewSell).Methods("POST")
	router.HandleFunc("/new/order", NewOrder).Methods("POST")

	log.Fatal(http.ListenAndServe(":23000", router))

}

func NewInvest(w http.ResponseWriter, r *http.Request) {

	var input BuyFund
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		var info InfoMessage
		info.Code = 300
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.ID == "" {
		var info InfoMessage
		info.Code = 301
		info.Message = "Investor ID can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.Fund == "" {
		var info InfoMessage
		info.Code = 302
		info.Message = "Fund can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.Amount == 0 {
		var info InfoMessage
		info.Code = 303
		info.Message = "Amount can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	conn, err := amqp.Dial(RabbitConnect)
	if err != nil {
		conn.Close()
		var info InfoMessage
		info.Code = 500
		info.Message = "Failed to connect to RabbitMQ"
		json.NewEncoder(w).Encode(info)
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		ch.Close()
		var info InfoMessage
		info.Code = 500
		info.Message = "Failed to open a channel"
		json.NewEncoder(w).Encode(info)
		return
	}

	q, err := ch.QueueDeclare(
		BuyQueue, // name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = "Failed to declare a queue"
		json.NewEncoder(w).Encode(info)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, _ := json.Marshal(input)

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		})

	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = "Failed to publish a message"
		json.NewEncoder(w).Encode(info)
		return
	} else {
		var info InfoMessage
		info.Code = 200
		info.Message = "A message is published"
		json.NewEncoder(w).Encode(info)
		return
	}

}

func NewSell(w http.ResponseWriter, r *http.Request) {

	var input Sell
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		var info InfoMessage
		info.Code = 300
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.ID == "" {
		var info InfoMessage
		info.Code = 301
		info.Message = "Investor ID can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.Fund == "" {
		var info InfoMessage
		info.Code = 302
		info.Message = "Fund can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.Qty == 0 {
		var info InfoMessage
		info.Code = 303
		info.Message = "Qty can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	conn, err := amqp.Dial(RabbitConnect)
	if err != nil {
		conn.Close()
		var info InfoMessage
		info.Code = 500
		info.Message = "Failed to connect to RabbitMQ"
		json.NewEncoder(w).Encode(info)
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		ch.Close()
		var info InfoMessage
		info.Code = 500
		info.Message = "Failed to open a channel"
		json.NewEncoder(w).Encode(info)
		return
	}

	q, err := ch.QueueDeclare(
		SellQueue, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = "Failed to declare a queue"
		json.NewEncoder(w).Encode(info)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, _ := json.Marshal(input)

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		})

	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = "Failed to publish a message"
		json.NewEncoder(w).Encode(info)
		return
	} else {
		var info InfoMessage
		info.Code = 200
		info.Message = "A message is published"
		json.NewEncoder(w).Encode(info)
		return
	}

}

func NewOrder(w http.ResponseWriter, r *http.Request) {

	var input Order
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		var info InfoMessage
		info.Code = 300
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.ID == "" {
		var info InfoMessage
		info.Code = 301
		info.Message = "Order ID can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.Investor == "" {
		var info InfoMessage
		info.Code = 302
		info.Message = "Investor can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.Fund == "" {
		var info InfoMessage
		info.Code = 303
		info.Message = "Fund can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.Qty == 0 {
		var info InfoMessage
		info.Code = 304
		info.Message = "Qty can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	conn, err := amqp.Dial(RabbitConnect)
	if err != nil {
		conn.Close()
		var info InfoMessage
		info.Code = 500
		info.Message = "Failed to connect to RabbitMQ"
		json.NewEncoder(w).Encode(info)
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		ch.Close()
		var info InfoMessage
		info.Code = 500
		info.Message = "Failed to open a channel"
		json.NewEncoder(w).Encode(info)
		return
	}

	q, err := ch.QueueDeclare(
		OrderQueue, // name
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = "Failed to declare a queue"
		json.NewEncoder(w).Encode(info)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, _ := json.Marshal(input)

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		})

	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = "Failed to publish a message"
		json.NewEncoder(w).Encode(info)
		return
	} else {
		var info InfoMessage
		info.Code = 200
		info.Message = "A message is published"
		json.NewEncoder(w).Encode(info)
		return
	}

}
