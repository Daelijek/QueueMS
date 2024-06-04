package messaging

import (
    "log"

    "github.com/streadway/amqp"
)

func SubscribeToQueue() {
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %s", err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        log.Fatalf("Failed to open a channel: %s", err)
    }
    defer ch.Close()

    q, err := ch.QueueDeclare(
        "client_queue", // name
        false,          // durable
        false,          // delete when unused
        false,          // exclusive
        false,          // no-wait
        nil,            // arguments
    )
    if err != nil {
        log.Fatalf("Failed to declare a queue: %s", err)
    }

    msgs, err := ch.Consume(
        q.Name, // queue
        "",     // consumer
        true,   // auto-ack
        false,  // exclusive
        false,  // no-local
        false,  // no-wait
        nil,    // args
    )
    if err != nil {
        log.Fatalf("Failed to register a consumer: %s", err)
    }

    go func() {
        for d := range msgs {
            log.Printf("Received a message: %s", d.Body)
            // Process the message
        }
    }()

    log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
    select {}
}
