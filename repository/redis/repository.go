package redis

import (
	"fmt"
	"hex-microservice/shortener"
	"strconv"

	"github.com/go-redis/redis"
)

const (
	key_code      = shortener.RedirectKeyCode
	key_url       = shortener.RedirectKeyURL
	key_createdAt = shortener.RedirectKeyCreatedAt
)

type redisRepository struct {
	client *redis.Client
}

func newClient(URL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(URL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	if _, err := client.Ping().Result(); err != nil {
		return nil, err
	}

	return client, nil
}

func New(url string) (shortener.Repository, error) {
	client, err := newClient(url)
	if err != nil {
		return nil, fmt.Errorf("repository.NewRedisRepository: %w", err)
	}

	return &redisRepository{
		client: client,
	}, nil
}

func generateKey(code string) string {
	return fmt.Sprintf("redirect:%s", code)
}

func (r *redisRepository) Find(code string) (*shortener.Redirect, error) {
	key := generateKey(code)
	data, err := r.client.HGetAll(key).Result()
	if err != nil {
		return nil, fmt.Errorf("repository.Redirect.Find: %w", err)
	}
	if len(data) == 0 {
		return nil, shortener.ErrRedirectNotFound
	}

	createdAt, err := strconv.ParseInt(data[key_createdAt], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("repository.Redirect.Find: %w", err)
	}

	return &shortener.Redirect{
		Code:      data[key_code],
		URL:       data[key_url],
		CreatedAt: createdAt,
	}, nil
}

func (r *redisRepository) Store(redirect *shortener.Redirect) error {
	key := generateKey(redirect.Code)

	data := map[string]interface{}{
		key_code:      redirect.Code,
		key_url:       redirect.URL,
		key_createdAt: redirect.CreatedAt,
	}

	if _, err := r.client.HMSet(key, data).Result(); err != nil {
		return fmt.Errorf("repository.Redirect.Store: %w", err)
	}

	return nil
}
