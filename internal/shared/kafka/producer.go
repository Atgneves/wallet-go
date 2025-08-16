package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: writer,
	}, nil
}

func (p *Producer) SendMessage(ctx context.Context, topic string, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	}

	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		log.Printf("Failed to send message to topic %s: %v", topic, err)
		return err
	}

	log.Printf("Message successfully sent to topic: %s", topic)
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
