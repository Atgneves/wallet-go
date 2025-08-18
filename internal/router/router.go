package router

import (
	"wallet-go/internal/health"
	"wallet-go/internal/operation"
	"wallet-go/internal/shared/config"
	"wallet-go/internal/shared/database"
	"wallet-go/internal/shared/kafka"
	"wallet-go/internal/shared/middleware"
	"wallet-go/internal/shared/utils"
	"wallet-go/internal/wallet"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(mongoClient *database.MongoClient, kafkaProducer *kafka.Producer, cfg *config.Config) *gin.Engine {
	// Set Gin mode
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Middlewares
	r.Use(gin.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.CORS())

	// Initialize dependencies
	lockManager := utils.NewWalletLockManager()

	// Stores
	walletStore := wallet.NewStore(mongoClient)
	operationStore := operation.NewStore(mongoClient)

	// Validators
	walletValidator := wallet.NewValidator()

	// Services
	walletService := wallet.NewService(walletStore, operationStore, walletValidator, lockManager)
	operationService := operation.NewService(operationStore)
	healthService := health.NewService(mongoClient, []string{cfg.Kafka.Brokers[0]}, cfg)

	// Handlers
	walletHandler := wallet.NewHandler(walletService, operationService, kafkaProducer, cfg.Kafka.Topics.Deposit, cfg.Kafka.Topics.Withdraw, cfg.Kafka.Topics.Transfer)
	operationHandler := operation.NewHandler(operationService)
	healthHandler := health.NewHandler(healthService)

	// Swagger route (before another routes)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Redirect root to swagger
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})

	// API Routes
	setupWalletRoutes(r, walletHandler, operationHandler)
	setupOperationRoutes(r, operationHandler)
	setupHealthRoutes(r, healthHandler)

	return r
}

func setupWalletRoutes(r *gin.Engine, walletHandler *wallet.Handler, operationHandler *operation.Handler) {
	walletGroup := r.Group("/wallet")
	{
		walletGroup.POST("", walletHandler.Create)
		walletGroup.GET("", walletHandler.List)
		walletGroup.GET("/:id", walletHandler.GetByID)
		walletGroup.PATCH("/:id", walletHandler.Patch)
		walletGroup.POST("/:id/deposit", walletHandler.Deposit)
		walletGroup.POST("/:id/withdraw", walletHandler.Withdraw)
		walletGroup.POST("/:id/transfer", walletHandler.Transfer)

		// Rotas de operation movidas para dentro do grupo wallet
		walletGroup.GET("/daily-summary", operationHandler.GetDailySummary)
		walletGroup.GET("/daily-summary-details", operationHandler.GetDailySummaryDetails)

	}
}

func setupOperationRoutes(r *gin.Engine, operationHandler *operation.Handler) {
	operationGroup := r.Group("/operations")
	{
		operationGroup.GET("", operationHandler.List)
		operationGroup.GET("/:operationId", operationHandler.GetByID)
	}
}

func setupHealthRoutes(r *gin.Engine, healthHandler *health.Handler) {
	healthGroup := r.Group("/health")
	{
		healthGroup.GET("", healthHandler.Health)
		healthGroup.GET("/details", healthHandler.HealthDetails)
	}
}
