package mongo

import (
	"context"
	"errors"
	"fmt"
	"hex-microservice/shortener"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

const (
	key_code      = shortener.RedirectKeyCode
	key_url       = shortener.RedirectKeyURL
	key_createdAt = shortener.RedirectKeyCreatedAt

	defaultDatabase   = "shortener"
	defaultCollection = "shortener"
	defaultTimeout    = 1 * time.Minute

	option_collection = "collection"
	option_timeout    = "timeout"
)

type connection struct {
	database   string
	collection string
	timeout    time.Duration
}

type mongoRepository struct {
	*connection
	client *mongo.Client
}

func connectionFromURL(URL string) (*connection, error) {
	// mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]

	c, err := connstring.ParseAndValidate(URL)
	if err != nil {
		return nil, err
	}

	con := &connection{
		database:   defaultDatabase,
		collection: defaultCollection,
		timeout:    defaultTimeout,
	}

	// database
	if c.Database != "" {
		con.database = c.Database
	}

	// collection
	col, ok := c.UnknownOptions[option_collection]
	if ok {
		con.collection = strings.Join(col, "")
	}

	// timeout
	// TODO: find out `opts.ConnectTimeout` should/can be used instead
	timeout, ok := c.UnknownOptions[option_timeout]
	if ok {
		timeoutAsDuration, err := time.ParseDuration(strings.Join(timeout, ""))
		if err != nil {
			return nil, fmt.Errorf("repository.NewMongoRepository: can't pase timeout: %w", err)
		}

		con.timeout = timeoutAsDuration
	}

	return con, nil
}

func newClient(URL string, timeout time.Duration) (*mongo.Client, error) {
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
func New(url string) (shortener.Repository, error) {
	con, err := connectionFromURL(url)
	if err != nil {
		return nil, err
	}

	client, err := newClient(url, con.timeout)
	if err != nil {
		return nil, fmt.Errorf("repository.NewMongoRepo: %w", err)
	}

	return &mongoRepository{
		connection: con,
		client:     client,
	}, nil
}

func (r *mongoRepository) Find(code string) (*shortener.Redirect, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	redirect := &shortener.Redirect{}
	collection := r.client.Database(r.database).Collection(r.collection)
	filter := bson.M{key_code: code}

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
			key_code:      redirect.Code,
			key_url:       redirect.URL,
			key_createdAt: redirect.CreatedAt,
		},
	); err != nil {
		return fmt.Errorf("repository.Redirect.Store: %w", err)
	}

	return nil
}
