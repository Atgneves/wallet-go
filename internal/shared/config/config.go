package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server  ServerConfig
	MongoDB MongoDBConfig
	Kafka   KafkaConfig
	Health  HealthConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type MongoDBConfig struct {
	URI      string
	Database string
}

type KafkaConfig struct {
	Brokers []string
	GroupID string
	Topics  KafkaTopics
}

type KafkaTopics struct {
	Deposit  string
	Withdraw string
	Transfer string
}

type HealthConfig struct {
	ShowDetails bool
}

// TODO: adjust to use replics primary, secundary and secundary2

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "wallet"),
		},
		Kafka: KafkaConfig{
			Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:29092")},
			GroupID: getEnv("KAFKA_GROUP_ID", "wallet-group"),
			Topics: KafkaTopics{
				Deposit:  getEnv("KAFKA_TOPIC_DEPOSIT", "wallet.deposit"),
				Withdraw: getEnv("KAFKA_TOPIC_WITHDRAW", "wallet.withdraw"),
				Transfer: getEnv("KAFKA_TOPIC_TRANSFER", "wallet.transfer"),
			},
		},
		Health: HealthConfig{
			ShowDetails: getBoolEnv("HEALTH_SHOW_DETAILS", false),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return defaultValue
}
