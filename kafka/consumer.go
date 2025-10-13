package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/tair/full-observability/pkg/logger"
)

// Consumer wraps Kafka consumer
type Consumer struct {
	consumer      sarama.ConsumerGroup
	brokers       []string
	groupID       string
	topics        []string
	handlers      map[string]EventHandler
	handlersMutex sync.RWMutex
}

// EventHandler is a function that handles events
type EventHandler func(ctx context.Context, event ProductPurchasedEvent) error

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, groupID string, topics []string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	logger.Logger.Info().
		Strs("brokers", brokers).
		Str("group_id", groupID).
		Strs("topics", topics).
		Msg("Kafka consumer initialized")

	return &Consumer{
		consumer: consumer,
		brokers:  brokers,
		groupID:  groupID,
		topics:   topics,
		handlers: make(map[string]EventHandler),
	}, nil
}

// RegisterHandler registers an event handler for a specific event type
func (c *Consumer) RegisterHandler(eventType string, handler EventHandler) {
	c.handlersMutex.Lock()
	defer c.handlersMutex.Unlock()
	c.handlers[eventType] = handler
	logger.Logger.Info().
		Str("event_type", eventType).
		Msg("Event handler registered")
}

// Start starts consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	handler := &consumerGroupHandler{
		consumer: c,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Logger.Info().Msg("Consumer context cancelled, stopping...")
				return
			default:
				if err := c.consumer.Consume(ctx, c.topics, handler); err != nil {
					logger.Logger.Error().
						Err(err).
						Msg("Error from consumer")
				}
			}
		}
	}()

	// Handle errors
	go func() {
		for err := range c.consumer.Errors() {
			logger.Logger.Error().
				Err(err).
				Msg("Consumer error")
		}
	}()

	logger.Logger.Info().
		Strs("topics", c.topics).
		Str("group_id", c.groupID).
		Msg("Kafka consumer started")

	return nil
}

// Close closes the Kafka consumer
func (c *Consumer) Close() error {
	if c.consumer != nil {
		return c.consumer.Close()
	}
	return nil
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	consumer *Consumer
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.handleMessage(session.Context(), message)
		session.MarkMessage(message, "")
	}
	return nil
}

func (h *consumerGroupHandler) handleMessage(ctx context.Context, message *sarama.ConsumerMessage) {
	// Extract trace context from Kafka headers
	carrier := propagation.MapCarrier{}
	for _, header := range message.Headers {
		key := string(header.Key)
		// Only extract trace context headers
		if key == "traceparent" || key == "tracestate" {
			carrier[key] = string(header.Value)
		}
	}

	// Extract context with trace
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	// Start consumer span
	tracer := otel.Tracer("kafka-consumer")
	ctx, span := tracer.Start(ctx, "kafka.consume.product_purchased",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.source", message.Topic),
			attribute.String("messaging.source_kind", "topic"),
			attribute.Int("messaging.kafka.partition", int(message.Partition)),
			attribute.Int64("messaging.kafka.offset", message.Offset),
		),
	)
	defer span.End()

	logger.Logger.Debug().
		Str("topic", message.Topic).
		Int32("partition", message.Partition).
		Int64("offset", message.Offset).
		Str("trace_id", span.SpanContext().TraceID().String()).
		Msg("Received message")

	// Get event type from headers
	eventType := ""
	eventID := ""
	for _, header := range message.Headers {
		if string(header.Key) == "event_type" {
			eventType = string(header.Value)
		}
		if string(header.Key) == "event_id" {
			eventID = string(header.Value)
		}
	}

	if eventType == "" {
		span.SetStatus(codes.Error, "Message without event_type header")
		logger.Logger.Warn().Msg("Message without event_type header")
		return
	}

	span.SetAttributes(
		attribute.String("event.type", eventType),
		attribute.String("event.id", eventID),
	)

	// Get handler for event type
	h.consumer.handlersMutex.RLock()
	handler, exists := h.consumer.handlers[eventType]
	h.consumer.handlersMutex.RUnlock()

	if !exists {
		span.SetStatus(codes.Error, "No handler registered")
		logger.Logger.Warn().
			Str("event_type", eventType).
			Msg("No handler registered for event type")
		return
	}

	// Parse event based on type
	switch eventType {
	case EventTypeProductPurchased:
		var event ProductPurchasedEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to unmarshal event")
			logger.Logger.Error().
				Err(err).
				Str("event_type", eventType).
				Msg("Failed to unmarshal event")
			return
		}

		span.SetAttributes(
			attribute.Int64("product.id", int64(event.ProductID)),
			attribute.Int("product.quantity", int(event.Quantity)),
			attribute.Int64("payment.id", int64(event.PaymentID)),
		)

		// Handle event
		if err := handler(ctx, event); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to handle event")
			logger.Logger.Error().
				Err(err).
				Str("event_type", eventType).
				Str("event_id", event.EventID).
				Str("trace_id", span.SpanContext().TraceID().String()).
				Msg("Failed to handle event")
			return
		}

		span.SetStatus(codes.Ok, "Event handled successfully")
		logger.Logger.Info().
			Str("event_type", eventType).
			Str("event_id", event.EventID).
			Uint("product_id", event.ProductID).
			Int32("quantity", event.Quantity).
			Str("trace_id", span.SpanContext().TraceID().String()).
			Msg("Event handled successfully")

	default:
		span.SetStatus(codes.Error, "Unknown event type")
		logger.Logger.Warn().
			Str("event_type", eventType).
			Msg("Unknown event type")
	}
}

