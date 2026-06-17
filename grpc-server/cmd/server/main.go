package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "grpc-server/proto"
	"google.golang.org/grpc"
)

// Estructura que implementa la interfaz generada por Protobuf
type server struct {
	pb.UnimplementedMatchPredictionServiceServer
}

// Función que se ejecuta cuando el Cliente hace la llamada
func (s *server) SendPrediction(ctx context.Context, req *pb.MatchPredictionRequest) (*pb.MatchPredictionResponse, error) {
	log.Printf("[GRPC SERVER] Predicción recibida: %v vs %v (Goles: %d - %d) por %s", 
		req.HomeTeam, req.AwayTeam, req.HomeGoals, req.AwayGoals, req.Username)

	// TODO: En el próximo paso, aquí enviaremos el mensaje a RabbitMQ.

	return &pb.MatchPredictionResponse{Status: "Recibido en gRPC Server exitosamente"}, nil
}

func main() {
	// Escuchamos en el puerto 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Fallo al escuchar: %v", err)
	}

	// Creamos la instancia del servidor gRPC
	s := grpc.NewServer()

	// Registramos nuestro servicio en el servidor
	pb.RegisterMatchPredictionServiceServer(s, &server{})

	fmt.Println(" Servidor gRPC (Microservicio 3) escuchando en puerto :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Fallo al servir: %v", err)
	}
}
