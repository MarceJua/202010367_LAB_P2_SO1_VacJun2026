package main

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Conexión a Valkey (Local)
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Fallo al conectar a Valkey: %v", err)
	}
	log.Println("Conectado a Valkey exitosamente")

	// 2. Conexión a RabbitMQ (Local)
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Fallo al conectar a RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Fallo al abrir el canal de RabbitMQ: %v", err)
	}
	defer ch.Close()

	// 3. Declarar la cola (por si el consumer arranca antes que el server)
	q, err := ch.QueueDeclare("quiniela_queue", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Fallo al declarar la cola: %v", err)
	}

	// 4. Empezar a consumir mensajes
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack (el mensaje se borra de la cola al leerse)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Fallo al registrar el consumidor: %v", err)
	}

	fmt.Println("Consumer Go (Microservicio 4) esperando mensajes...")

	// Ciclo infinito escuchando mensajes
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf("[RABBITMQ CONSUMER] Mensaje recibido: %s", d.Body)

			// Guardar el mensaje en Valkey (en una lista llamada "predicciones")
			err := rdb.LPush(ctx, "predicciones", d.Body).Err()
			if err != nil {
				log.Printf("[ERROR] No se pudo guardar en Valkey: %v", err)
			} else {
				log.Println("💾 Guardado en Valkey exitosamente")
			}
		}
	}()
	<-forever
}
