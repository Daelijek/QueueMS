package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/Daelijek/queue-management-system/client-service/models"
    "github.com/Daelijek/queue-management-system/client-service/services"
)

func ClientHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        clients := services.GetAllClients()
        json.NewEncoder(w).Encode(clients)
    case "POST":
        var client models.Client
        json.NewDecoder(r.Body).Decode(&client)
        services.CreateClient(client)
        w.WriteHeader(http.StatusCreated)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}
