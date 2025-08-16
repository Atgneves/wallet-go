package health

import (
	"context"
	"time"

	"wallet-go/internal/shared/config"
	"wallet-go/internal/shared/database"
	"github.com/segmentio/kafka-go"
)

type Service struct {
	mongoClient  *database.MongoClient
	kafkaBrokers []string
	config       *config.Config
}

func NewService(mongoClient *database.MongoClient, kafkaBrokers []string, config *config.Config) *Service {
	return &Service{
		mongoClient:  mongoClient,
		kafkaBrokers: kafkaBrokers,
		config:       config,
	}
}

func (s *Service) GetHealth() *Health {
	health := s.GetHealthWithDetails()

	if !s.config.Health.ShowDetails {
		if health.Status == HealthStatusDown {
			return &Health{Status: HealthStatusDown}
		}
		return &Health{Status: HealthStatusUp}
	}

	return health
}

func (s *Service) GetHealthWithDetails() *Health {
	details := make(map[string]interface{})

	mongoHealth := s.checkMongoHealth()
	kafkaHealth := s.checkKafkaHealth()

	details["mongodb"] = mongoHealth
	details["kafka"] = kafkaHealth

	overallStatus := HealthStatusUp
	if mongoHealth.Status == HealthStatusDown || kafkaHealth.Status == HealthStatusDown {
		overallStatus = HealthStatusDown
	}

	return &Health{
		Status:  overallStatus,
		Details: details,
	}
}

func (s *Service) checkMongoHealth() *ComponentHealth {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.mongoClient.Client.Ping(ctx, nil)
	if err != nil {
		return &ComponentHealth{
			Status: HealthStatusDown,
			Error:  err.Error(),
		}
	}

	return &ComponentHealth{
		Status: HealthStatusUp,
	}
}

func (s *Service) checkKafkaHealth() *ComponentHealth {
	conn, err := kafka.DialContext(context.Background(), "tcp", s.kafkaBrokers[0])
	if err != nil {
		return &ComponentHealth{
			Status: HealthStatusDown,
			Error:  err.Error(),
		}
	}
	defer conn.Close()

	return &ComponentHealth{
		Status: HealthStatusUp,
	}
}
