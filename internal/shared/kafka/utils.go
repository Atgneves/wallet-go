package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// CreateTopics cria os tópicos Kafka necessários para a aplicação
func CreateTopics(brokers []string, topics []string) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafka.Dial("tcp", controller.Host+":"+string(rune(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfigs := make([]kafka.TopicConfig, len(topics))
	for i, topic := range topics {
		topicConfigs[i] = kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
	}

	return controllerConn.CreateTopics(topicConfigs...)
}

// WaitForKafka aguarda o Kafka ficar disponível com retry
func WaitForKafka(brokers []string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		conn, err := kafka.DialContext(context.Background(), "tcp", brokers[0])
		if err == nil {
			conn.Close()
			log.Println("Kafka is ready")
			return nil
		}

		log.Printf("Waiting for Kafka... attempt %d/%d", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}

	// Última tentativa
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
