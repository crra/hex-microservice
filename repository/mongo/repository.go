package mongo

import (
	"context"
	"errors"
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/lookup"
	"hex-microservice/meta/value"
	"hex-microservice/repository"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

const (
	key_code = "code"

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
	connection
	client *mongo.Client
	parent context.Context
}

func connectionFromURL(URL string) (connection, error) {
	// mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]
	var con connection

	c, err := connstring.ParseAndValidate(URL)
	if err != nil {
		return con, err
	}

	con = connection{
		collection: defaultCollection,
		timeout:    defaultTimeout,
		database:   value.OrDefault(&c.Database, defaultDatabase),
	}

	// collection
	// this is so much shorter and clearer than:
	//    strings.Join(
	//      value.GetFromMap(c.UnknownOptions,
	//		  value.PointerOf(option_collection),
	//		  value.Ident[string],
	//		  []string{defaultCollection},
	//	  ), "")
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
			return con, fmt.Errorf("repository.NewMongoRepository: can't pase timeout: %w", err)
		}

		con.timeout = timeoutAsDuration
	}

	return con, nil
}

func newClient(parent context.Context, URL string, timeout time.Duration) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	// note: connstring.ParseAndValidate is called again during `applyURI`
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
func New(parent context.Context, url string) (repository.RedirectRepository, error) {
	con, err := connectionFromURL(url)
	if err != nil {
		return nil, err
	}

	client, err := newClient(parent, url, con.timeout)
	if err != nil {
		return nil, fmt.Errorf("repository.NewMongoRepo: %w", err)
	}

	return &mongoRepository{
		connection: con,
		client:     client,
		parent:     parent,
	}, nil
}

func (r *mongoRepository) LookupFind(code string) (lookup.RedirectStorage, error) {
	var red lookup.RedirectStorage
	ctx, cancel := context.WithTimeout(r.parent, r.timeout)
	defer cancel()

	stored := &redirect{}
	collection := r.client.Database(r.database).Collection(r.collection)
	filter := bson.M{key_code: code}

	if err := collection.FindOne(ctx, filter).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return red, lookup.ErrNotFound
		}

		return red, fmt.Errorf("repository.Redirect.Find: %w", err)
	}

	// From from database domain to application domain
	// TODO: replace with generated
	if err := copier.Copy(&red, &stored); err != nil {
		return red, fmt.Errorf("repository.Redirect.Find copying: %w", err)
	}

	return red, nil
}

func (r *mongoRepository) Store(red adder.RedirectStorage) error {
	ctx, cancel := context.WithTimeout(r.parent, r.timeout)
	defer cancel()

	collection := r.client.Database(r.database).Collection(r.collection)

	var store redirect

	// From from application domain to database domain
	// TODO: replace with generated
	if err := copier.Copy(&store, &red); err != nil {
		return fmt.Errorf("repository.Redirect.Store copying: %w", err)
	}

	if _, err := collection.InsertOne(ctx, store); err != nil {
		return fmt.Errorf("repository.Redirect.Store: %w", err)
	}

	return nil
}

//////////////////////////////////////////////////////////////////

func (r *mongoRepository) DeleteFind(code string) (deleter.RedirectStorage, error) {
	var red deleter.RedirectStorage

	return red, nil
}

func (r *mongoRepository) Delete(code, token string) error {
	return nil
}
