package middleware

import (
	"log"

	"wallet-go/internal/shared/errors"
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			log.Printf("Error processing request: %v", err)

			switch e := err.Err.(type) {
			case *errors.AppError:
				c.JSON(e.Code, e)
			default:
				appErr := errors.InternalServerError("Internal server error")
				c.JSON(appErr.Code, appErr)
			}

			c.Abort()
		}
	}
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("Panic recovered: %v", recovered)

		appErr := errors.InternalServerError("Internal server error")
		c.JSON(appErr.Code, appErr)
		c.Abort()
	})
}
