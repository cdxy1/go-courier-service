package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/cdxy1/go-courier-service/internal/model"
	"github.com/Shopify/sarama"
)

type OrderEventHandler interface {
	Handle(ctx context.Context, event model.OrderStatusEvent) error
}

type Consumer struct {
	group   sarama.ConsumerGroup
	topic   string
	handler OrderEventHandler
	ready   chan struct{}
}

func NewConsumer(brokers []string, groupID, topic, version string, handler OrderEventHandler) (*Consumer, error) {
	if len(brokers) == 0 {
		return nil, errInvalidConfig("brokers are empty")
	}
	if groupID == "" {
		return nil, errInvalidConfig("group id is empty")
	}
	if topic == "" {
		return nil, errInvalidConfig("topic is empty")
	}
	if handler == nil {
		return nil, errInvalidConfig("handler is nil")
	}

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_1_0_0

	if version != "" {
		parsed, err := sarama.ParseKafkaVersion(version)
		if err != nil {
			return nil, err
		}
		config.Version = parsed
	}

	group, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		group:   group,
		topic:   topic,
		handler: handler,
		ready:   make(chan struct{}),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	go func() {
		for err := range c.group.Errors() {
			log.Printf("kafka consumer error: %v", err)
		}
	}()

	for {
		if err := c.group.Consume(ctx, []string{c.topic}, c); err != nil {
			return err
		}
		if ctx.Err() != nil {
			return nil
		}
		c.ready = make(chan struct{})
	}
}

func (c *Consumer) Close() error {
	return c.group.Close()
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if message == nil {
			continue
		}
		var event model.OrderStatusEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("kafka event decode failed: %v", err)
			session.MarkMessage(message, "decode failed")
			continue
		}
		if err := c.handler.Handle(session.Context(), event); err != nil {
			log.Printf("kafka event handling failed: %v", err)
		}
		session.MarkMessage(message, "")
	}
	return nil
}

type errInvalidConfig string

func (e errInvalidConfig) Error() string {
	return string(e)
}
