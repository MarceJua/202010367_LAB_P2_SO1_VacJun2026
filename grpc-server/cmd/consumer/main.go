package main

import (
	"context"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)
func main() {
	// Variables de entorno
	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "localhost:6379"
	}

	rabbitUrl := os.Getenv("RABBITMQ_URL")
	if rabbitUrl == "" {
		rabbitUrl = "amqp://guest:guest@localhost:5672/"
	}

	// 1. Conexión a Valkey
	rdb := redis.NewClient(&redis.Options{Addr: valkeyAddr})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Fallo al conectar a Valkey en %s: %v", valkeyAddr, err)
	}
	log.Println("Conectado a Valkey exitosamente")

	// 2. Conexión a RabbitMQ
	conn, err := amqp.Dial(rabbitUrl)
	if err != nil {
		log.Fatalf("Fallo al conectar a RabbitMQ en %s: %v", rabbitUrl, err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Fallo al abrir el canal de RabbitMQ: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("quiniela_queue", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Fallo al declarar la cola: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Fallo al registrar el consumidor: %v", err)
	}

	fmt.Println("Consumer Go esperando mensajes...")

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			err := rdb.LPush(ctx, "predicciones", d.Body).Err()
			if err != nil {
				log.Printf("[ERROR] No se pudo guardar en Valkey: %v", err)
			} else {
				log.Println("Guardado en Valkey exitosamente")
			}
		}
	}()
	<-forever
}
