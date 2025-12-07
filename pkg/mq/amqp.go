package mq

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ExtractVhostFromURL extracts the vhost from an AMQP URL.
// Examples:
// - amqp://user:pass@host:5672/myvhost -> "myvhost"
// - amqp://user:pass@host:5672/ -> "/"
// - amqp://user:pass@host:5672 -> "/"
func ExtractVhostFromURL(amqpURL string) (string, error) {
	if amqpURL == "" {
		return "", fmt.Errorf("empty url")
	}
	parsed, err := url.Parse(amqpURL)
	if err != nil {
		return "", err
	}
	// Only accept amqp/amqps schemes
	if parsed.Scheme != "amqp" && parsed.Scheme != "amqps" {
		return "", fmt.Errorf("unsupported scheme: %s", parsed.Scheme)
	}
	// URL-decode the path to handle encoded vhosts
	decoded, err := url.PathUnescape(parsed.Path)
	if err != nil {
		// fallback to raw path
		decoded = parsed.Path
	}
	if decoded == "" || decoded == "/" {
		return "/", nil
	}
	// Remove a single leading slash. If there are two ("//..."), the result will keep one leading slash,
	// which matches the expectation for encoded vhosts (e.g., "/%2Fmyvhost" -> "/myvhost").
	return strings.TrimPrefix(decoded, "/"), nil
}

// RabbitMQClient handles the connection to RabbitMQ
type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

// NewRabbitMQClient creates a new RabbitMQ client
func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	client := &RabbitMQClient{
		url: url,
	}

	if err := client.connect(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *RabbitMQClient) connect() error {
	var err error

	log.Printf("Connecting to RabbitMQ...")
	// Don't log the full URL if it contains credentials
	c.conn, err = amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	log.Println("Successfully connected to RabbitMQ")
	return nil
}

// GetChannel returns the AMQP channel
func (c *RabbitMQClient) GetChannel() *amqp.Channel {
	return c.channel
}

// Close closes the connection and channel
func (c *RabbitMQClient) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// EnsureExchange ensures the exchange exists
func (c *RabbitMQClient) EnsureExchange(name, kind string) error {
	return c.channel.ExchangeDeclare(
		name,  // name
		kind,  // type
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
}
