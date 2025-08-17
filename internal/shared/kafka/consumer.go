package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

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
	log.Println("Starting Kafka consumers...")
	for topic, reader := range c.readers {
		log.Printf("Starting consumer for topic: %s", topic)
		go c.consumeMessages(topic, reader)
		time.Sleep(100 * time.Millisecond) // Pequeno delay entre consumers
	}
}

func (c *Consumer) consumeMessages(topic string, reader *kafka.Reader) {
	for {
		log.Printf("Waiting for message on topic: %s", topic)

		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message from topic %s: %v", topic, err)
			continue
		}

		log.Printf("Received message on topic %s: %s", topic, string(message.Value))

		if err := c.processMessage(topic, message.Value); err != nil {
			log.Printf("Error processing message from topic %s: %v", topic, err)
		} else {
			log.Printf("Successfully processed message from topic %s", topic)
		}
	}
}

func (c *Consumer) processMessage(topic string, data []byte) error {
	log.Printf("Processing message for topic: %s, data: %s", topic, string(data))

	ctx := context.Background()

	switch topic {
	case "wallet.deposit":
		log.Println("Processing deposit...")
		var msg WalletKafkaTransactionMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Error unmarshaling deposit message: %v", err)
			return err
		}
		log.Printf("Calling DepositFromKafka with walletID: %s, amount: %d", msg.WalletID, msg.AmountInCents)
		return c.walletService.DepositFromKafka(ctx, msg.WalletID, msg.AmountInCents)

	case "wallet.withdraw":
		log.Println("Processing withdraw...")
		var msg WalletKafkaTransactionMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Error unmarshaling withdraw message: %v", err)
			return err
		}
		log.Printf("Calling WithdrawFromKafka with walletID: %s, amount: %d", msg.WalletID, msg.AmountInCents)
		return c.walletService.WithdrawFromKafka(ctx, msg.WalletID, msg.AmountInCents)

	case "wallet.transfer":
		log.Println("Processing transfer...")
		var msg WalletKafkaTransactionTransferMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Error unmarshaling transfer message: %v", err)
			return err
		}
		log.Printf("Calling TransferFromKafka with sourceID: %s, amount: %d, destinationID: %s",
			msg.WalletID, msg.AmountInCents, msg.WalletDestinationID)
		return c.walletService.TransferFromKafka(ctx, msg.WalletID, msg.AmountInCents, msg.WalletDestinationID)

	default:
		log.Printf("Unknown topic: %s", topic)
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
