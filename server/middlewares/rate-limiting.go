package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
)

func NewUserRateLimiter(logger logrus.FieldLogger, limit string) gin.HandlerFunc {
	if limit != "" {
		rate, err := limiter.NewRateFromFormatted(limit)
		if err != nil {
			logger.Errorf("error getting rate limit: %s\n", err)
		}
		return mgin.NewMiddleware(limiter.New(memory.NewStore(), rate),
			mgin.WithLimitReachedHandler(func(c *gin.Context) {
				logger.Warnf("Rate limit exceeded")
				c.JSON(http.StatusTooManyRequests, gin.H{"code": http.StatusTooManyRequests, "cause": "TooManyRequests"})
			}),

			mgin.WithErrorHandler(func(c *gin.Context, err error) {
				logger.Errorf("Rate limit error: %s\n", err)
			}),
		)
	}
	return func(c *gin.Context) {
		c.Next()
	}
}
