package events

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type EventManager struct {
	client *redis.Client
}

func NewEventManager(addr string) (*EventManager, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "",
		DB:           0,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &EventManager{client: rdb}, nil
}

func (em *EventManager) PublishEvent(ctx context.Context, channel string, payload interface{}) error {
	return em.client.Publish(ctx, channel, payload).Err()
}

func (em *EventManager) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return em.client.Subscribe(ctx, channel)
}

func (em *EventManager) Close() error {
	return em.client.Close()
}
