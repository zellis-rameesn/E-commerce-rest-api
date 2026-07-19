package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zellis-rameesn/go-ecommerce/internal/providers"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"

	_ "github.com/aws/smithy-go/endpoints"
	appconfig "github.com/zellis-rameesn/go-ecommerce/internal/config"
)

type EventPublisher struct {
	publisher message.Publisher
	queueName string
}

func NewEventPublisher(ctx context.Context, cfg appconfig.AWSConfig) (*EventPublisher, error) {
	logger := watermill.NewStdLogger(false, false)

	awsCfg, err := providers.CreateAwsConfig(ctx, cfg.Region, cfg.S3Endpoint, cfg.AccessKeyID, cfg.SecretAccessKey, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS config: %w", err)
	}

	publisherConfig := sqs.PublisherConfig{
		AWSConfig: awsCfg,
		Marshaler: nil,
	}

	publisher, err := sqs.NewPublisher(publisherConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	return &EventPublisher{
		publisher: publisher,
		queueName: cfg.EventQueueName,
	}, nil
}

func (ep *EventPublisher) Publish(eventType string, payload interface{}, metadata map[string]string) error {

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), data)
	msg.Metadata.Set("eventType", eventType)
	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	return ep.publisher.Publish(ep.queueName, msg)
}

func (ep *EventPublisher) Close() error {
	return ep.publisher.Close()
}
