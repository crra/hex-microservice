package redis

import (
	"context"
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/lookup"
	"hex-microservice/meta/value"
	"hex-microservice/repository"
	"time"

	redis "github.com/go-redis/redis/v8"
)

const redisStructAnnoationTag = "redis"

type redisRepository struct {
	client *redis.Client
	parent context.Context
	// list of keys that are stored as hash value in redis
	redisMappingsOfRedirect map[string]string
}

func newClient(parent context.Context, URL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	if _, err := client.Ping(parent).Result(); err != nil {
		return nil, err
	}

	return client, nil
}

func New(parent context.Context, url string) (repository.RedirectRepository, error) {
	client, err := newClient(parent, url)
	if err != nil {
		return nil, fmt.Errorf("redisRepository.New client creation: %w", err)
	}

	redisMappingsOfRedirect, err := value.Mapping(redirect{}, redisStructAnnoationTag)
	if err != nil {
		return nil, fmt.Errorf("redisRepository.New parsing redirect declaration: %w", err)
	}

	return &redisRepository{
		parent: parent,
		client: client,

		redisMappingsOfRedirect: redisMappingsOfRedirect,
	}, nil
}

func generateKey(code string) string {
	return fmt.Sprintf("redirect:%s", code)
}

func (r *redisRepository) LookupFind(code string) (lookup.RedirectStorage, error) {
	var red lookup.RedirectStorage
	key := generateKey(code)
	get := r.client.HMGet(r.parent, key, value.Keys(r.redisMappingsOfRedirect)...)

	values, err := get.Result()
	if err != nil {
		return red, fmt.Errorf("repository.Redirect.Find result: %w", err)
	}

	// like in: Array.prototype.some, at least one value is not nil
	notFound := true
	for _, v := range values {
		if v != nil {
			notFound = false
			break
		}
	}

	if notFound {
		return red, lookup.ErrNotFound
	}

	var stored redirect
	if err := get.Scan(&stored); err != nil {
		return red, fmt.Errorf("repository.RedirectFind scaning: %w", err)
	}

	stored.CreatedAt, err = time.Parse(time.RFC3339, stored.CreatedAtStr)
	if err != nil {
		return red, fmt.Errorf("repository.Redirect.Find parsing time: %w", err)
	}

	return fromRedirectToLookupRedirectStorage(stored), nil
}

func (r *redisRepository) Store(red adder.RedirectStorage) error {
	// From from application domain to database domain
	// TODO: replace with generated
	store := fromAdderRedirectStorageToRedirect(red)

	_, err := r.client.HSet(r.parent, red.Code,
		store.marshalHash(value.Keys(r.redisMappingsOfRedirect))).Result()
	return err
}

//////////////////////////////////////////////////////////////////

func (r *redisRepository) DeleteFind(code string) (deleter.RedirectStorage, error) {
	var red deleter.RedirectStorage

	return red, nil
}

func (r *redisRepository) Delete(code, token string) error {
	return nil
}
