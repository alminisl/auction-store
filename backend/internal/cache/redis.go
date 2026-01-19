package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *RedisCache) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// Pub/Sub for real-time bid updates
func (c *RedisCache) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return c.client.Publish(ctx, channel, data).Err()
}

func (c *RedisCache) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return c.client.Subscribe(ctx, channel)
}

// Auction-specific channel naming
func AuctionChannel(auctionID uuid.UUID) string {
	return fmt.Sprintf("auction:%s", auctionID.String())
}

// Rate limiting
func (c *RedisCache) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	pipe := c.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

func (c *RedisCache) GetRateLimit(ctx context.Context, key string) (int64, error) {
	val, err := c.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// Rate limit key generators
func RateLimitKeyIP(ip string) string {
	return fmt.Sprintf("ratelimit:ip:%s", ip)
}

func RateLimitKeyUser(userID uuid.UUID) string {
	return fmt.Sprintf("ratelimit:user:%s", userID.String())
}

func RateLimitKeyAuth(ip string) string {
	return fmt.Sprintf("ratelimit:auth:%s", ip)
}

func RateLimitKeyBid(userID uuid.UUID) string {
	return fmt.Sprintf("ratelimit:bid:%s", userID.String())
}

// Client returns the underlying redis client for advanced operations
func (c *RedisCache) Client() *redis.Client {
	return c.client
}
