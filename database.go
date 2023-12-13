package mongosteps

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// DatabaseOption sets the database option.
type DatabaseOption interface {
	applyDatabaseOption(d *database)
}

type databaseOptionFunc func(d *database)

func (f databaseOptionFunc) applyDatabaseOption(d *database) {
	f(d)
}

type database struct {
	conn     *mongo.Database
	cleanUps []string
}

// cleanUp cleans up the collections in the database.
func (d *database) cleanUp(ctx context.Context) error {
	for _, collection := range d.cleanUps {
		if err := d.truncate(ctx, collection); err != nil {
			return err
		}
	}

	return nil
}

// find returns the documents in the collection that match the filter.
func (d *database) find(ctx context.Context, collection string, filter interface{}, opts ...*options.FindOptions) ([]bsoncore.Document, error) {
	cursor, err := d.conn.Collection(collection).Find(ctx, filter, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not find documents in collection %q: %w", collection, err)
	}

	defer cursor.Close(ctx) // nolint: errcheck

	var result []bsoncore.Document

	if err := cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("could not read documents from collection %q: %w", collection, err)
	}

	if len(result) == 0 {
		return []bsoncore.Document{}, nil
	}

	return result, nil
}

// truncate deletes all the documents the collection.
func (d *database) truncate(ctx context.Context, collection string) error {
	if _, err := d.conn.Collection(collection).DeleteMany(ctx, bson.D{}); err != nil {
		return fmt.Errorf("could not truncate collection %q: %w", collection, err)
	}

	return nil
}

func (d *database) store(ctx context.Context, collection string, docs []bsoncore.Document) error {
	documents := make([]interface{}, len(docs))
	for i, doc := range docs {
		documents[i] = doc
	}

	if _, err := d.conn.Collection(collection).InsertMany(ctx, documents); err != nil {
		return fmt.Errorf("could not insert documents into collection %q: %w", collection, err)
	}

	return nil
}

func (d *database) count(ctx context.Context, collection string, filter interface{}) (int64, error) {
	count, err := d.conn.Collection(collection).CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("could not count documents in collection %q: %w", collection, err)
	}

	return count, nil
}

// newDatabase creates a new database.
func newDatabase(conn *mongo.Database, opts ...DatabaseOption) *database {
	d := &database{
		conn: conn,
	}

	for _, opt := range opts {
		opt.applyDatabaseOption(d)
	}

	return d
}

// CleanUpAfterScenario cleans up the collections in the database after the scenario.
func CleanUpAfterScenario(collections ...string) DatabaseOption {
	return databaseOptionFunc(func(d *database) {
		d.cleanUps = append(d.cleanUps, collections...)
	})
}
