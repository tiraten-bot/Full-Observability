package kafka

import "time"

// ProductPurchasedEvent represents a product purchase event
type ProductPurchasedEvent struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	PaymentID     uint      `json:"payment_id"`
	ProductID     uint      `json:"product_id"`
	Quantity      int32     `json:"quantity"`
	UserID        uint      `json:"user_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	PaymentMethod string    `json:"payment_method"`
	Timestamp     time.Time `json:"timestamp"`
}

// Event types
const (
	EventTypeProductPurchased = "product.purchased"
)

// Kafka topics
const (
	TopicProductPurchased = "product-purchased"
)
