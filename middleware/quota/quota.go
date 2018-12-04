package quota

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

type (
	// API Quota middleware interface definition.
	APIQuotaInterface interface {
		// Parse token from http request.
		ParseTokenFromRequest(ctx context.Context, r *http.Request) (string, error)
		// Check current token's quota plan.
		GetTokenQuotaPlan(ctx context.Context, token string) (int64, error)
	}

	// API Quota middleware package initiation options.
	APIQuotaOption struct {
		KeyPrefix string
		Cache     *redis.Pool
		Logger    *logrus.Logger
		Interface APIQuotaInterface
	}
)

var (
	keyPrefix string
	cache     *redis.Pool
	logger    *logrus.Logger
	apiQuota  APIQuotaInterface
)

// Initiate package's variables and dependencies.
func Init(opt APIQuotaOption) {
	keyPrefix = opt.KeyPrefix
	cache = opt.Cache
	logger = opt.Logger
	apiQuota = opt.Interface
}

// Get cache key name on given api key.
func CacheKey(token string) string {
	return fmt.Sprintf("api-quota:%s:%s", keyPrefix, token)
}

// Get total request / usage on given api key.
func getUsage(ctx context.Context, token string) (int64, error) {
	conn, err := cache.GetContext(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	return redis.Int64(conn.Do("GET", CacheKey(token)))
}

// Increment total request / usage on given api key.
func incUsage(ctx context.Context, token string) (int64, error) {
	conn, err := cache.GetContext(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	return redis.Int64(conn.Do("INCR", CacheKey(token)))
}

// HTTP Middleware for api quota request limiter.
func HTTPMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get api key / token which represents a client.
		token, err := apiQuota.ParseTokenFromRequest(ctx, r)
		if err != nil {
			logger.Warnln("Error parse token from request:", err)
			http.Error(w, "missing api token", http.StatusBadRequest)
			return
		}

		// Check current total usage in redis by api key.
		usageCount, err := getUsage(ctx, token)
		if err != nil {
			logger.Warnf("Error get usage count for token '%s': %+v\n", token, err)

			if err == redis.ErrNil {
				http.Error(w, "invalid api token", http.StatusUnauthorized)
				return
			}

			http.Error(w, "unknown error", http.StatusInternalServerError)
			return
		}

		// Validate current total api usage.
		planQuota, err := apiQuota.GetTokenQuotaPlan(ctx, token)
		if err != nil {
			logger.Warnf("Error get token plan quota for token '%s': %+v\n", token, err)
			http.Error(w, "unknown error", http.StatusInternalServerError)
			return
		}

		if usageCount >= planQuota && planQuota != -1 {
			http.Error(w, "quota limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Increment api total usage.
		go func() {
			_, err := incUsage(ctx, token)
			if err != nil {
				logger.Warnf("Error update token usage for token '%s': %+v\n", token, err)
			}
		}()

		next.ServeHTTP(w, r)
	}
}
