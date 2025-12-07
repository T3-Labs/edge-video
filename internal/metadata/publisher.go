package metadata

import (
	"encoding/json"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type EventType string

const (
	EventTypeFrame        EventType = "frame"
	EventTypeCameraStatus EventType = "camera_status"
	EventTypeSystemStatus EventType = "system_status"
)

type CameraState string

const (
	CameraStateActive   CameraState = "active"
	CameraStateInactive CameraState = "inactive"
	CameraStateOffline  CameraState = "offline"
)

// Publisher handles publishing frame metadata to RabbitMQ.
type Publisher struct {
	channel    *amqp.Channel
	exchange   string
	routingKey string
	ttlSeconds int
}

// NewPublisher creates a new metadata Publisher.
func NewPublisher(ch *amqp.Channel, exchange, routingKey string, ttlSeconds int) *Publisher {
	if ch != nil {
		err := ch.ExchangeDeclare(
			exchange,
			"topic",
			true,  // durable
			false, // auto-deleted
			false, // internal
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			// Log o erro mas não falha a criação do publisher
			// Isso permite que a aplicação continue mesmo se o exchange já existir
			// ou houver problemas temporários
		}
	}

	return &Publisher{
		channel:    ch,
		exchange:   exchange,
		routingKey: routingKey,
		ttlSeconds: ttlSeconds,
	}
}

// Metadata represents the structure of the metadata message.
type Metadata struct {
	EventType     EventType `json:"event_type"`
	CameraID      string    `json:"camera_id"`
	Timestamp     time.Time `json:"timestamp"`
	TimestampNano int64     `json:"timestamp_nano"`
	Sequence      string    `json:"sequence,omitempty"`
	RedisKey      string    `json:"redis_key,omitempty"`
	Width         int       `json:"width,omitempty"`
	Height        int       `json:"height,omitempty"`
	Encoding      string    `json:"encoding,omitempty"`
	SizeBytes     int       `json:"size_bytes,omitempty"`
	TTLSeconds    int       `json:"ttl_seconds"`
}

type CameraStatusEvent struct {
	EventType           EventType   `json:"event_type"`
	CameraID            string      `json:"camera_id"`
	Timestamp           time.Time   `json:"timestamp"`
	State               CameraState `json:"state"`
	ConsecutiveFailures int         `json:"consecutive_failures,omitempty"`
	LastError           string      `json:"last_error,omitempty"`
	Message             string      `json:"message,omitempty"`
}

type SystemStatusEvent struct {
	EventType       EventType `json:"event_type"`
	Timestamp       time.Time `json:"timestamp"`
	TotalCameras    int       `json:"total_cameras"`
	ActiveCameras   int       `json:"active_cameras"`
	InactiveCameras int       `json:"inactive_cameras"`
	Message         string    `json:"message"`
}

// PublishMetadata sends a JSON message with frame metadata to RabbitMQ.
func (p *Publisher) PublishMetadata(cameraID string, timestamp time.Time, redisKey string, width, height, size int, encoding string) error {
	metadata := Metadata{
		EventType:     EventTypeFrame,
		CameraID:      cameraID,
		Timestamp:     timestamp,
		TimestampNano: timestamp.UnixNano(),
		RedisKey:      redisKey,
		Width:         width,
		Height:        height,
		Encoding:      encoding,
		SizeBytes:     size,
		TTLSeconds:    p.ttlSeconds,
	}

	body, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	expiration := ""
	if p.ttlSeconds > 0 {
		// RabbitMQ expects expiration as milliseconds string
		expiration = strconv.FormatInt(int64(p.ttlSeconds*1000), 10)
	}

	return p.channel.Publish(
		p.exchange,
		p.routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Expiration:  expiration,
		},
	)
}

// PublishCameraStatus sends camera status change events to RabbitMQ.
func (p *Publisher) PublishCameraStatus(cameraID string, state CameraState, consecutiveFailures int, lastError error, message string) error {
	event := CameraStatusEvent{
		EventType:           EventTypeCameraStatus,
		CameraID:            cameraID,
		Timestamp:           time.Now(),
		State:               state,
		ConsecutiveFailures: consecutiveFailures,
		Message:             message,
	}

	if lastError != nil {
		event.LastError = lastError.Error()
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.channel.Publish(
		p.exchange,
		p.routingKey+".status",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

// PublishSystemStatus sends system-wide status events to RabbitMQ.
func (p *Publisher) PublishSystemStatus(totalCameras, activeCameras, inactiveCameras int, message string) error {
	event := SystemStatusEvent{
		EventType:       EventTypeSystemStatus,
		Timestamp:       time.Now(),
		TotalCameras:    totalCameras,
		ActiveCameras:   activeCameras,
		InactiveCameras: inactiveCameras,
		Message:         message,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.channel.Publish(
		p.exchange,
		p.routingKey+".system",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
