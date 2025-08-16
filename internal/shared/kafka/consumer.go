package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// WalletService interface simplificada para quebrar dependência circular
type WalletService interface {
	DepositFromKafka(ctx context.Context, walletID uuid.UUID, amountInCents int64) error
	WithdrawFromKafka(ctx context.Context, walletID uuid.UUID, amountInCents int64) error
	TransferFromKafka(ctx context.Context, sourceID uuid.UUID, amountInCents int64, destinationID uuid.UUID) error
}

// WalletKafkaTransactionMessage representa mensagem de transação simples
type WalletKafkaTransactionMessage struct {
	WalletID      uuid.UUID `json:"walletId"`
	AmountInCents int64     `json:"amountInCents"`
}

// WalletKafkaTransactionTransferMessage representa mensagem de transferência
type WalletKafkaTransactionTransferMessage struct {
	WalletID            uuid.UUID `json:"walletId"`
	AmountInCents       int64     `json:"amountInCents"`
	WalletDestinationID uuid.UUID `json:"walletDestinationId"`
}

type Consumer struct {
	readers       map[string]*kafka.Reader
	walletService WalletService
}

func NewConsumer(brokers []string, groupID string) (*Consumer, error) {
	readers := make(map[string]*kafka.Reader)

	// Create readers for each topic
	topics := []string{"wallet.deposit", "wallet.withdraw", "wallet.transfer"}
	for _, topic := range topics {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		})
		readers[topic] = reader
	}

	return &Consumer{
		readers: readers,
	}, nil
}

func (c *Consumer) SetWalletService(service WalletService) {
	c.walletService = service
}

func (c *Consumer) StartConsumers() {
	for topic, reader := range c.readers {
		go c.consumeMessages(topic, reader)
	}
}

func (c *Consumer) consumeMessages(topic string, reader *kafka.Reader) {
	for {
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message from topic %s: %v", topic, err)
			continue
		}

		if err := c.processMessage(topic, message.Value); err != nil {
			log.Printf("Error processing message from topic %s: %v", topic, err)
		}
	}
}

func (c *Consumer) processMessage(topic string, data []byte) error {
	ctx := context.Background()

	switch topic {
	case "wallet.deposit":
		var msg WalletKafkaTransactionMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return err
		}

		return c.walletService.DepositFromKafka(ctx, msg.WalletID, msg.AmountInCents)

	case "wallet.withdraw":
		var msg WalletKafkaTransactionMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return err
		}

		return c.walletService.WithdrawFromKafka(ctx, msg.WalletID, msg.AmountInCents)

	case "wallet.transfer":
		var msg WalletKafkaTransactionTransferMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return err
		}

		return c.walletService.TransferFromKafka(ctx, msg.WalletID, msg.AmountInCents, msg.WalletDestinationID)
	}

	return nil
}

func (c *Consumer) Close() error {
	for _, reader := range c.readers {
		if err := reader.Close(); err != nil {
			log.Printf("Error closing kafka reader: %v", err)
		}
	}
	return nil
}
