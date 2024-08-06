package internal

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	client *redis.Client
	key    string
}

func NewRateLimiter(client *redis.Client, key string) *RateLimiter {
	return &RateLimiter{
		client: client,
		key:    key,
	}
}

func (r *RateLimiter) Allow(ctx context.Context, key string, value string) (bool, error) {
	var rateLimit int
	var err error

	rateLimit, err = strconv.Atoi(os.Getenv("RATE_LIMITER_IP"))
	if err != nil {
		return false, fmt.Errorf("invalid RATE_LIMITER_IP: %v", err)
	}

	if key == "token" {
		rateLimit, err = strconv.Atoi(os.Getenv("RATE_LIMITER_TOKEN"))
		if err != nil {
			return false, fmt.Errorf("invalid RATE_LIMITER_TOKEN: %v", err)
		}
	}

	windowDuration := time.Second * time.Duration(rateLimit)

	uid := fmt.Sprintf("%s:%d", value, time.Now().Unix())

	counter, err := r.client.Get(ctx, uid).Int()
	if err == redis.Nil {
		err = r.client.Set(ctx, uid, 1, windowDuration).Err()
		if err != nil {
			return false, fmt.Errorf("failed to set initial value: %v", err)
		}
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to get value: %v", err)
	}

	if counter >= rateLimit {
		return false, nil
	}

	_, err = r.client.Incr(ctx, uid).Result()
	if err != nil {
		return false, fmt.Errorf("failed to increment value: %v", err)
	}

	return true, nil
}
