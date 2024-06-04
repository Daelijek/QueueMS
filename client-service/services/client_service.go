package services

import (
    "github.com/Daelijek/queue-management-system/client-service/models"
)

var clients = []models.Client{}

func GetAllClients() []models.Client {
    return clients
}

func CreateClient(client models.Client) {
    clients = append(clients, client)
}
