package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/rs/zerolog"
)

type Consumer struct {
	readers  []*kafka.Reader
	handlers map[string]EventHandler
	logger   zerolog.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

type EventHandler func(ctx context.Context, message kafka.Message) error

func NewConsumer(brokers []string, groupID string, topics []string, logger zerolog.Logger) *Consumer {
	ctx, cancel := context.WithCancel(context.Background())

	readers := make([]*kafka.Reader, 0, len(topics))
	for _, topic := range topics {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			GroupID:        groupID,
			Topic:          topic,
			MinBytes:       10e3,
			MaxBytes:       10e6,
			MaxWait:        1 * time.Second,
			CommitInterval: 1 * time.Second,
			StartOffset:    kafka.LastOffset,
		})
		readers = append(readers, reader)
	}

	return &Consumer{
		readers:  readers,
		handlers: make(map[string]EventHandler),
		logger:   logger.With().Str("component", "kafka").Logger(),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (c *Consumer) RegisterHandler(topic string, handler EventHandler) {
	c.handlers[topic] = handler
}

func (c *Consumer) Start() {
	for _, reader := range c.readers {
		go c.consumeFromReader(reader)
	}
	c.logger.Info().Int("topics", len(c.readers)).Msg("Kafka consumer started")
}

func (c *Consumer) consumeFromReader(reader *kafka.Reader) {
	topic := reader.Config().Topic
	c.logger.Info().Str("topic", topic).Msg("Starting consumer for topic")

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg, err := reader.FetchMessage(c.ctx)
			if err != nil {
				if c.ctx.Err() != nil {
					return
				}
				c.logger.Error().Err(err).Str("topic", topic).Msg("Failed to fetch message")
				time.Sleep(1 * time.Second)
				continue
			}

			c.logger.Debug().
				Str("topic", topic).
				Int("partition", msg.Partition).
				Int64("offset", msg.Offset).
				Msg("Received message")

			handler, ok := c.handlers[topic]
			if !ok {
				c.logger.Warn().Str("topic", topic).Msg("No handler registered for topic")
				reader.CommitMessages(c.ctx, msg)
				continue
			}

			if err := handler(c.ctx, msg); err != nil {
				c.logger.Error().Err(err).Str("topic", topic).Msg("Handler failed")
			}

			if err := reader.CommitMessages(c.ctx, msg); err != nil {
				c.logger.Error().Err(err).Str("topic", topic).Msg("Failed to commit message")
			}
		}
	}
}

func (c *Consumer) Stop() error {
	c.cancel()

	var lastErr error
	for _, reader := range c.readers {
		if err := reader.Close(); err != nil {
			lastErr = err
			c.logger.Error().Err(err).Msg("Failed to close reader")
		}
	}

	c.logger.Info().Msg("Kafka consumer stopped")
	return lastErr
}
