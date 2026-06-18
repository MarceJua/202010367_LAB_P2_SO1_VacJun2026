package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"os"

	pb "grpc-server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Prediction struct {
	HomeTeam  string `json:"home_team"`
	AwayTeam  string `json:"away_team"`
	HomeGoals int32  `json:"home_goals"`
	AwayGoals int32  `json:"away_goals"`
	Username  string `json:"username"`
	Timestamp string `json:"timestamp"`
}

// Función auxiliar para convertir el string al Enum de Protobuf
func parseTeam(teamStr string) pb.Teams {
	switch teamStr {
	case "GTM": return pb.Teams_GTM
	case "MEX": return pb.Teams_MEX
	case "BRA": return pb.Teams_BRA
	case "ARG": return pb.Teams_ARG
	case "ESP": return pb.Teams_ESP
	default: return pb.Teams_TEAMS_UNKNOWN
	}
}

func predictionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var p Prediction
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Variable de entorno para el Servidor gRPC
	grpcAddr := os.Getenv("GRPC_SERVER_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:50051"
	}

	log.Printf("[GO CLIENT] Enviando por gRPC a %s...", grpcAddr)

	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar a gRPC: %v", err)
	}
	defer conn.Close()

	client := pb.NewMatchPredictionServiceClient(conn)

	req := &pb.MatchPredictionRequest{
		HomeTeam:  parseTeam(p.HomeTeam),
		AwayTeam:  parseTeam(p.AwayTeam),
		HomeGoals: p.HomeGoals,
		AwayGoals: p.AwayGoals,
		Username:  p.Username,
		Timestamp: p.Timestamp,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := client.SendPrediction(ctx, req)
	grpcStatus := "Error en comunicación gRPC"
	if err != nil {
		log.Printf("[ERROR] Fallo al llamar a gRPC: %v", err)
	} else {
		grpcStatus = res.GetStatus()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "Procesado por Go Client",
		"grpc_response": grpcStatus,
	})
}

func main() {
	http.HandleFunc("/", predictionHandler)
	fmt.Println("Cliente Go escuchando en puerto :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
