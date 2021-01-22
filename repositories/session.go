package repositories

import (
	"context"
	"encoding/json"
	"time"

	"../models"
	"github.com/go-redis/redis"
)

// SessionRepository contains db operations for type models.Session
type SessionRepository interface {
	Set(context.Context, string, models.Session, time.Duration) error
	Get(context.Context, string) (models.Session, error)
	Publish(context.Context, string, models.Session)
	Subscribe(context.Context, string) *redis.PubSub
}

type sessionRepository struct {
	Client *redis.Client
}

// NewSessionRepository returns a value representing the
// SessionRepository interface
func NewSessionRepository(client *redis.Client) SessionRepository {
	return &sessionRepository{
		Client: client,
	}
}

func (sr *sessionRepository) Get(ctx context.Context, key string) (models.Session, error) {
	result, err := sr.Client.Get(ctx, key).Result()

	if err != nil {
		return models.Session{}, err
	}

	var sess models.Session
	json.Unmarshal([]byte(result), &sess)
	return sess, nil
}

func (sr *sessionRepository) Set(
	ctx context.Context,
	key string,
	sess models.Session,
	expires time.Duration,
) error {
	return sr.Client.Set(ctx, key, sess, expires).Err()
}

func (sr *sessionRepository) Publish(
	ctx context.Context,
	channel string,
	sess models.Session,
) {
	b, _ := json.Marshal(sess)
	sr.Client.Publish(ctx, channel, string(b))
}

func (sr *sessionRepository) Subscribe(
	ctx context.Context,
	channel string,
) *redis.PubSub {
	ps := sr.Client.Subscribe(ctx, channel)
	// Wait for subscription notice
	ps.Receive(ctx)
	return ps
}
