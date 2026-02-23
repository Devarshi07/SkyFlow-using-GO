package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/skyflow/skyflow/internal/shared/logger"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	log  *logger.Logger
	mu   sync.Mutex
}

func NewClient(url string, log *logger.Logger) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq connect: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}
	return &Client{conn: conn, ch: ch, log: log}, nil
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// DeclareQueue declares a durable queue
func (c *Client) DeclareQueue(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.ch.QueueDeclare(name, true, false, false, false, nil)
	return err
}

// Publish sends a JSON message to the given queue
func (c *Client) Publish(ctx context.Context, queue string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ch.PublishWithContext(ctx,
		"",    // default exchange
		queue, // routing key = queue name
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
}

// Consume starts consuming messages from a queue and calls handler for each
func (c *Client) Consume(queue, consumerTag string, handler func([]byte) error) error {
	msgs, err := c.ch.Consume(queue, consumerTag, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("rabbitmq consume: %w", err)
	}
	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				c.log.Error("message handler failed", "queue", queue, "error", err)
				msg.Nack(false, true) // requeue
			} else {
				msg.Ack(false)
			}
		}
	}()
	return nil
}
