package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/zellis-rameesn/go-ecommerce/internal/config"
	"github.com/zellis-rameesn/go-ecommerce/internal/events"
	"github.com/zellis-rameesn/go-ecommerce/internal/models"
	"github.com/zellis-rameesn/go-ecommerce/internal/notifications"
	"github.com/zellis-rameesn/go-ecommerce/internal/providers"
)

func main() {
	log.Println("Started notification service")
	cfg := config.Load()
	notifier := notifications.NewEmailNotifier(&cfg.SMTP)

	ctx := context.Background()
	awsConfig, err := providers.CreateAwsConfig(ctx, cfg.AWS.Region, cfg.AWS.S3Endpoint, cfg.AWS.AccessKeyID, cfg.AWS.SecretAccessKey, "")
	if err != nil {
		log.Fatalf("Failed to create aws config: %v", err)
	}
	subConfig := sqs.SubscriberConfig{
		AWSConfig:   awsConfig,
		Unmarshaler: nil,
	}
	logger := watermill.NewStdLogger(false, false)

	subscriber, err := sqs.NewSubscriber(subConfig, logger)
	if err != nil {
		subscriber.Close()
		log.Fatalf("Failed to initialize subscriber: %v", err)
	}

	quit := make(chan os.Signal, 1)
	msg, err := subscriber.Subscribe(ctx, cfg.AWS.EventQueueName)

	if err != nil {
		subscriber.Close()
		log.Fatalf("Failed to subscribe to queue: %v", err)
	}

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case newMsg := <-msg:
			if err := processMessageAndSendNotification(newMsg, notifier); err != nil {
				log.Printf("Error processing message: %v", err)
				newMsg.Nack()
			} else {
				newMsg.Ack()
			}

		case <-quit:
			log.Println("Shutting down notification service")
			subscriber.Close()
			return
		}
	}

}

func processMessageAndSendNotification(msg *message.Message, notifier *notifications.EmailNotifier) error {
	var user models.User
	if err := json.Unmarshal([]byte(msg.Payload), &user); err != nil {
		log.Printf("Failed to unmarshal payload: %s", err.Error())
		return err
	}

	userName := user.FirstName + " " + user.LastName
	if userName == " " {
		userName = "User"
	}
	eventType := msg.Metadata.Get("eventType")

	switch eventType {
	case events.USER_LOGGED_IN:
		if err := notifier.SendLoginNotification(user.Email, userName); err != nil {
			log.Printf("Failed to send notification: %s", err.Error())
			return err
		}
	default:
		log.Printf("Unknown event type: %s", eventType)
		return errors.New("unknown event")
	}

	return nil
}
