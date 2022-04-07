package mongosteps

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cucumber/godog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func stringToDocs(data *godog.DocString) ([]bsoncore.Document, error) {
	if data == nil {
		return nil, errors.New("data is nil") // nolint: goerr113
	}

	return bytesToDocs([]byte(data.Content))
}

func bytesToDocs(data []byte) ([]bsoncore.Document, error) {
	var docs []bsoncore.Document

	if err := bson.UnmarshalExtJSON(data, true, &docs); err != nil {
		return nil, fmt.Errorf("error unmarshaling extjson: %w", err)
	}

	return docs, nil
}

func docsToExtJSON(docs []bsoncore.Document) ([]byte, error) {
	if len(docs) == 0 {
		return []byte("[]"), nil
	}

	result := make([]json.RawMessage, len(docs))

	for i, doc := range docs {
		result[i] = json.RawMessage(doc.String())
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("error marshaling documents: %w", err)
	}

	data = bytes.ReplaceAll(data, []byte(`\u003cignored-diff\u003e`), []byte("<ignore-diff>"))

	return data, nil
}

func stringToBSOND(data *godog.DocString) (bson.D, error) {
	if data == nil {
		return bson.D{}, nil
	}

	return bytesToBSOND([]byte(data.Content))
}

func bytesToBSOND(data []byte) (bson.D, error) {
	var result bson.D

	if err := bson.UnmarshalExtJSON(data, true, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling extjson: %w", err)
	}

	return result, nil
}
