package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

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

// PublishProductPurchased publishes a product purchased event with tracing
func (p *Publisher) PublishProductPurchased(ctx context.Context, event ProductPurchasedEvent) error {
	// Start tracing span
	tracer := otel.Tracer("kafka-publisher")
	ctx, span := tracer.Start(ctx, "kafka.publish.product_purchased",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", TopicProductPurchased),
			attribute.String("messaging.destination_kind", "topic"),
			attribute.String("event.type", EventTypeProductPurchased),
			attribute.Int64("product.id", int64(event.ProductID)),
			attribute.Int("product.quantity", int(event.Quantity)),
			attribute.Int64("payment.id", int64(event.PaymentID)),
		),
	)
	defer span.End()

	// Set event metadata
	if event.EventID == "" {
		event.EventID = fmt.Sprintf("evt_%d", time.Now().UnixNano())
	}
	event.EventType = EventTypeProductPurchased
	event.Timestamp = time.Now()

	span.SetAttributes(attribute.String("event.id", event.EventID))

	// Marshal event to JSON
	eventBytes, err := json.Marshal(event)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to marshal event")
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Inject trace context into Kafka headers
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	headers := []sarama.RecordHeader{
		{
			Key:   []byte("event_type"),
			Value: []byte(EventTypeProductPurchased),
		},
		{
			Key:   []byte("event_id"),
			Value: []byte(event.EventID),
		},
	}

	// Add trace context to headers
	for key, value := range carrier {
		headers = append(headers, sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(value),
		})
	}

	// Create Kafka message
	msg := &sarama.ProducerMessage{
		Topic:   TopicProductPurchased,
		Key:     sarama.StringEncoder(fmt.Sprintf("product_%d", event.ProductID)),
		Value:   sarama.ByteEncoder(eventBytes),
		Headers: headers,
	}

	// Send message
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to send message")
		logger.Logger.Error().
			Err(err).
			Str("topic", TopicProductPurchased).
			Uint("product_id", event.ProductID).
			Str("trace_id", span.SpanContext().TraceID().String()).
			Msg("Failed to publish event")
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	span.SetAttributes(
		attribute.Int("messaging.kafka.partition", int(partition)),
		attribute.Int64("messaging.kafka.offset", offset),
	)
	span.SetStatus(codes.Ok, "Event published successfully")

	logger.Logger.Info().
		Str("event_id", event.EventID).
		Str("event_type", event.EventType).
		Str("topic", TopicProductPurchased).
		Int32("partition", partition).
		Int64("offset", offset).
		Uint("product_id", event.ProductID).
		Int32("quantity", event.Quantity).
		Str("trace_id", span.SpanContext().TraceID().String()).
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

