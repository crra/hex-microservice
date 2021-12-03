package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"hex-microservice/shortener"
)

type mongoRepository struct {
	client     *mongo.Client
	database   string
	collection string
	timeout    time.Duration
}

func newMongoClient(URL string, timeout time.Duration) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URL))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}

// time.Duration(timeout) * time.Second
func NewMongoRepository(url, db string, timeout time.Duration) (shortener.RedirectRepository, error) {
	client, err := newMongoClient(url, timeout)
	if err != nil {
		return nil, fmt.Errorf("repository.NewMongoRepo: %w", err)
	}

	return &mongoRepository{
		timeout:    timeout,
		database:   db,
		collection: "redirects",
		client:     client,
	}, nil
}

func (r *mongoRepository) Find(code string) (*shortener.Redirect, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	redirect := &shortener.Redirect{}
	collection := r.client.Database(r.database).Collection(r.collection)
	filter := bson.M{"code": code}

	if err := collection.FindOne(ctx, filter).Decode(&redirect); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shortener.ErrRedirectNotFound
		}

		return nil, fmt.Errorf("repository.Redirect.Find: %w", err)
	}

	return redirect, nil
}

func (r *mongoRepository) Store(redirect *shortener.Redirect) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	collection := r.client.Database(r.database).Collection(r.collection)
	if _, err := collection.InsertOne(
		ctx,
		bson.M{
			"code":       redirect.Code,
			"url":        redirect.URL,
			"created_at": redirect.CreatedAt,
		},
	); err != nil {
		return fmt.Errorf("repository.Redirect.Store: %w", err)
	}

	return nil
}
