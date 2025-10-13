package kafka

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/tair/full-observability/pkg/logger"
)

// Publisher wraps Kafka producer
type Publisher struct {
	producer sarama.SyncProducer
	brokers  []string
}

// NewPublisher creates a new Kafka publisher
func NewPublisher(brokers []string) (*Publisher, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 3
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.MaxMessageBytes = 1000000

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	logger.Logger.Info().
		Strs("brokers", brokers).
		Msg("Kafka publisher initialized")

	return &Publisher{
		producer: producer,
		brokers:  brokers,
	}, nil
}

// PublishProductPurchased publishes a product purchased event
func (p *Publisher) PublishProductPurchased(event ProductPurchasedEvent) error {
	// Set event metadata
	if event.EventID == "" {
		event.EventID = fmt.Sprintf("evt_%d", time.Now().UnixNano())
	}
	event.EventType = EventTypeProductPurchased
	event.Timestamp = time.Now()

	// Marshal event to JSON
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	msg := &sarama.ProducerMessage{
		Topic: TopicProductPurchased,
		Key:   sarama.StringEncoder(fmt.Sprintf("product_%d", event.ProductID)),
		Value: sarama.ByteEncoder(eventBytes),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event_type"),
				Value: []byte(EventTypeProductPurchased),
			},
			{
				Key:   []byte("event_id"),
				Value: []byte(event.EventID),
			},
		},
	}

	// Send message
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("topic", TopicProductPurchased).
			Uint("product_id", event.ProductID).
			Msg("Failed to publish event")
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	logger.Logger.Info().
		Str("event_id", event.EventID).
		Str("event_type", event.EventType).
		Str("topic", TopicProductPurchased).
		Int32("partition", partition).
		Int64("offset", offset).
		Uint("product_id", event.ProductID).
		Int32("quantity", event.Quantity).
		Msg("Product purchased event published")

	return nil
}

// Close closes the Kafka producer
func (p *Publisher) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

