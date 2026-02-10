package main

import (
	"log"
	"net"

	"github.com/pasanAbeysekara/collaborative-editor/internal/config"
	"github.com/pasanAbeysekara/collaborative-editor/internal/grpchandlers"
	pb "github.com/pasanAbeysekara/collaborative-editor/proto"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	cfg := config.Load()

	// Start gRPC server
	grpcPort := "50053"
	go startGRPCServer(grpcPort)

	// Continue with RabbitMQ consumer
	conn, err := amqp091.Dial(cfg.RabbitMQ_URL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declare the exchange. This is where the document-service will publish messages.
	err = ch.ExchangeDeclare(
		"events", // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	// Declare the queue that will receive the messages.
	q, err := ch.QueueDeclare(
		"notifications", // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Bind the queue to the exchange with a "routing key".
	// We are interested in any event related to users.
	log.Printf("Binding queue %s to exchange %s with routing key %s", q.Name, "events", "user.*")
	err = ch.QueueBind(
		q.Name,    // queue name
		"user.*",  // routing key (e.g., "user.invited", "user.deleted")
		"events",  // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	// Start consuming messages from the queue.
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			// Here we would process the message. For now, we just log it.
			log.Printf(" [x] Received a message with routing key '%s': %s", d.RoutingKey, d.Body)
			// Example logic: parse JSON, send email, etc.
		}
	}()

	log.Printf(" [*] Notification service is waiting for messages. To exit press CTRL+C")
	<-forever
}

func startGRPCServer(port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	notificationService := &grpchandlers.NotificationServiceServer{}
	pb.RegisterNotificationServiceServer(grpcServer, notificationService)

	log.Printf("Starting notification-service gRPC on port %s...\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}