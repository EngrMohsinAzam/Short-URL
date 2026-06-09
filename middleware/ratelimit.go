package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func RateLimiter() gin.HandlerFunc {
	// Allow 10 requests per minute per IP
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  10,
	}

	// Store in memory (no Redis needed for now)
	store := memory.NewStore()
	instance := limiter.New(store, rate)

	return func(c *gin.Context) {
		// Check limit for this IP
		context, err := instance.Get(c.Request.Context(), c.ClientIP())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limiter error"})
			c.Abort()
			return
		}

		// If limit exceeded → block the request
		if context.Reached {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests — slow down!",
				"retry_after": "1 minute",
			})
			c.Abort()
			return
		}

		// Add helpful headers to response
		c.Header("X-RateLimit-Limit", "10")
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))

		c.Next()
	}
}
