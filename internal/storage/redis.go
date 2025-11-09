package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Wrap the go-redis client and provide our app's interface to Redis operations
type RedisClient struct {
	client *redis.Client
}

// remember the NewClient command takes addr, password, and db (int)
type RedisConfig struct {
	Addr     string //whatever local host I'm putting it on
	Password string //empty for no pw
	DB       int    //db number, default should be 0
	Timeout  time.Duration
}

// Creates a new Reids client with the given config
func NewRedisClient(config RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  config.Timeout,
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
	})

	//context will time out after config.Timeout time, and cancel is the provided cancel function for this context (should always call it)
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Reids: %w", err) //%w preserves the oginal error and adds context
	}

	return &RedisClient{
		client: client,
	}, nil
}

// Ping check if Redis is alive, return nil if succesful
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.Client().Ping(ctx).Err()
}

// We expose the underlying go-redis client since our other packages will have to work with these methods directly anyway
func (r *RedisClient) Client() *redis.Client {
	return r.client
}

// close redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}
