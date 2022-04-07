package mongosteps

import (
	"context"

	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type queryerCtxKey struct{}

func contextWithDocs(ctx context.Context, docs []bsoncore.Document) context.Context {
	return context.WithValue(ctx, queryerCtxKey{}, docs)
}

func docsFromContext(ctx context.Context) []bsoncore.Document {
	q, ok := ctx.Value(queryerCtxKey{}).([]bsoncore.Document)
	if !ok || q == nil {
		return nil
	}

	return q
}
