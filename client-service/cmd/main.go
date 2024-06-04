package main

import (
	"github.com/Daelijek/queue-management-system/client-service/config"
	"github.com/Daelijek/queue-management-system/client-service/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	cfg := config.LoadConfig()

	http.HandleFunc("/clients", handlers.ClientHandler)

	log.Printf("Starting client service on port %s\n", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
