package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	pb "grpc-server/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

// Estructura del servidor que ahora incluye el canal de RabbitMQ
type server struct {
	pb.UnimplementedMatchPredictionServiceServer
	rabbitChannel *amqp.Channel
}

// Estructura para el JSON que enviaremos a la cola
type RabbitMessage struct {
	HomeTeam  string `json:"home_team"`
	AwayTeam  string `json:"away_team"`
	HomeGoals int32  `json:"home_goals"`
	AwayGoals int32  `json:"away_goals"`
	Username  string `json:"username"`
	Timestamp string `json:"timestamp"`
}

func (s *server) SendPrediction(ctx context.Context, req *pb.MatchPredictionRequest) (*pb.MatchPredictionResponse, error) {
	log.Printf("[GRPC SERVER] Predicción recibida: %v vs %v por %s", req.HomeTeam, req.AwayTeam, req.Username)
	
	// Preparamos el mensaje para RabbitMQ
	msg := RabbitMessage{
		HomeTeam:  req.HomeTeam.String(),
		AwayTeam:  req.AwayTeam.String(),
		HomeGoals: req.HomeGoals,
		AwayGoals: req.AwayGoals,
		Username:  req.Username,
		Timestamp: req.Timestamp,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[ERROR] Fallo al convertir a JSON: %v", err)
		return &pb.MatchPredictionResponse{Status: "Error interno procesando mensaje"}, nil
	}

	// Publicamos en la cola de RabbitMQ
	err = s.rabbitChannel.PublishWithContext(ctx,
		"",               // exchange
		"quiniela_queue", // routing key (nombre de la cola)
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		log.Printf("[ERROR] Fallo al publicar en RabbitMQ: %v", err)
		return &pb.MatchPredictionResponse{Status: "Error enviando a RabbitMQ"}, nil
	}

	log.Println("[RABBITMQ WRITER] Mensaje encolado exitosamente")
	return &pb.MatchPredictionResponse{Status: "Recibido y encolado en RabbitMQ exitosamente"}, nil
}

func main() {
	// 1. Conexión a RabbitMQ Local
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Fallo al conectar a RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Fallo al abrir un canal: %v", err)
	}
	defer ch.Close()

	// Declaramos la cola para asegurarnos de que exista
	_, err = ch.QueueDeclare(
		"quiniela_queue", // nombre
		true,             // durable (sobrevive a reinicios)
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Fatalf("Fallo al declarar la cola: %v", err)
	}

	// 2. Levantar servidor gRPC
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Fallo al escuchar: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMatchPredictionServiceServer(s, &server{rabbitChannel: ch})

	fmt.Println("Servidor gRPC + RabbitMQ Writer escuchando en puerto :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Fallo al servir: %v", err)
	}
}
