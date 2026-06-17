package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Prediction struct {
	HomeTeam  string `json:"home_team"`
	AwayTeam  string `json:"away_team"`
	HomeGoals int32  `json:"home_goals"`
	AwayGoals int32  `json:"away_goals"`
	Username  string `json:"username"`
	Timestamp string `json:"timestamp"`
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

	log.Printf("[GO CLIENT] JSON recibido desde Rust: %+v\n", p)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "Recibido exitosamente en Go Client",
	})
}

func main() {
	http.HandleFunc("/", predictionHandler)

	fmt.Println(" Cliente Go (REST) escuchando en http://localhost:8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}
