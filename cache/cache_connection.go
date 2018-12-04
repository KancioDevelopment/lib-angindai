package cache

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

type CacheOptions struct {
	Host        string
	Port        int
	Password    string
	MaxIdle     int
	MaxActive   int
	IdleTimeout int
	Enabled     bool
}

var pool *redis.Pool

func Connect(cacheOptions CacheOptions) *redis.Pool {
	if pool == nil {
		pool = &redis.Pool{
			MaxIdle:     cacheOptions.MaxIdle,
			MaxActive:   cacheOptions.MaxActive,
			IdleTimeout: time.Duration(cacheOptions.IdleTimeout) * time.Second,
			Dial: func() (redis.Conn, error) {
				address := fmt.Sprintf("%s:%d", cacheOptions.Host, cacheOptions.Port)
				c, err := redis.Dial("tcp", address)
				if err != nil {
					return nil, err
				}

				// Do authentication process if password not empty.
				if cacheOptions.Password != "" {
					if _, err := c.Do("AUTH", cacheOptions.Password); err != nil {
						c.Close()
						return nil, err
					}
				}

				return c, nil
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < time.Minute {
					return nil
				}

				_, err := c.Do("PING")
				return err
			},
			Wait:            true,
			MaxConnLifetime: 15 * time.Minute,
		}

		return pool
	}

	return pool
}
