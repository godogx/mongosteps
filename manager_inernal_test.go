package mongosteps

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestManager_NoDocumentsInCollectionOfDatabase(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		database      string
		result        int
		expectedError string
	}{
		{
			scenario:      "missing database",
			database:      "other",
			expectedError: `mongo database "other" is not registered to the manager`,
		},
		{
			scenario:      "truncate error",
			database:      defaultDatabase,
			expectedError: `could not truncate collection "customer": command failed`,
		},
		{
			scenario: "success",
			database: defaultDatabase,
			result:   1,
		},
	}

	for _, tc := range testCases {
		tc := tc

		mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

		mt.Cleanup(func() {
			mt.Close()
		})

		mt.Run(tc.scenario, func(t *mtest.T) {
			t.Parallel()

			t.AddMockResponses(bson.D{{Key: "ok", Value: tc.result}})

			m := NewManager(WithDefaultDatabase(t.DB))

			_, err := m.noDocumentsInCollectionOfDatabase(context.Background(), "customer", tc.database)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestManager_TheseDocumentsAreStoredInCollectionOfDatabase(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		database      string
		data          *godog.DocString
		result        int
		expectedError string
	}{
		{
			scenario:      "missing database",
			database:      "other",
			expectedError: `mongo database "other" is not registered to the manager`,
		},
		{
			scenario:      "no data",
			database:      defaultDatabase,
			expectedError: `failed to parse documents: data is nil`,
		},
		{
			scenario:      "malformed data",
			database:      defaultDatabase,
			data:          &godog.DocString{Content: "malformed"},
			expectedError: `failed to parse documents: error unmarshaling extjson: invalid JSON input`,
		},
		{
			scenario:      "empty docs",
			database:      defaultDatabase,
			data:          &godog.DocString{Content: "[]"},
			result:        0,
			expectedError: `could not insert documents into collection "customer": must provide at least one element in input slice`,
		},
		{
			scenario:      "store error",
			database:      defaultDatabase,
			data:          &godog.DocString{Content: `[{"message": "hello world"}]`},
			result:        0,
			expectedError: `could not insert documents into collection "customer": command failed`,
		},
		{
			scenario: "success",
			database: defaultDatabase,
			data:     &godog.DocString{Content: `[{"message": "hello world"}]`},
			result:   1,
		},
	}

	for _, tc := range testCases {
		tc := tc

		mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

		mt.Cleanup(func() {
			mt.Close()
		})

		mt.Run(tc.scenario, func(t *mtest.T) {
			t.Parallel()

			t.AddMockResponses(bson.D{{Key: "ok", Value: tc.result}})

			m := NewManager(WithDefaultDatabase(t.DB))

			_, err := m.theseDocumentsAreStoredInCollectionOfDatabase(context.Background(), "customer", tc.database, tc.data)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestManager_TheseDocumentsFromFileAreStoredInCollectionOfDatabase(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		database      string
		filePath      string
		result        int
		expectedError string
	}{
		{
			scenario:      "missing database",
			database:      "other",
			expectedError: `mongo database "other" is not registered to the manager`,
		},
		{
			scenario:      "file not found",
			database:      defaultDatabase,
			filePath:      "resources/fixtures/unknown.json",
			expectedError: `open resources/fixtures/unknown.json: no such file or directory`,
		},
		{
			scenario:      "malformed data",
			database:      defaultDatabase,
			filePath:      "resources/fixtures/malformed.json",
			expectedError: `error unmarshaling extjson: invalid JSON input; unexpected end of input at position 0`,
		},
		{
			scenario:      "empty docs",
			database:      defaultDatabase,
			filePath:      "resources/fixtures/empty.json",
			expectedError: `could not insert documents into collection "customer": must provide at least one element in input slice`,
		},
		{
			scenario:      "store error",
			database:      defaultDatabase,
			filePath:      "resources/fixtures/empty.json",
			expectedError: `could not insert documents into collection "customer": must provide at least one element in input slice`,
		},
		{
			scenario: "success",
			database: defaultDatabase,
			filePath: "resources/fixtures/customers.json",
			result:   1,
		},
	}

	for _, tc := range testCases {
		tc := tc

		mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

		mt.Cleanup(func() {
			mt.Close()
		})

		mt.Run(tc.scenario, func(t *mtest.T) {
			t.Parallel()

			t.AddMockResponses(bson.D{{Key: "ok", Value: tc.result}})

			m := NewManager(WithDefaultDatabase(t.DB))

			_, err := m.theseDocumentsFromFileAreStoredInCollectionOfDatabase(context.Background(), tc.filePath, "customer", tc.database)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestManager_SearchInCollectionOfDatabase(t *testing.T) {
	t.Parallel()

	docs := mustParseDocs(readFixtures("resources/fixtures/customers.json"))

	testCases := []struct {
		scenario        string
		database        string
		filter          *godog.DocString
		result          []bson.D
		expectedContext context.Context // nolint: containedctx
		expectedError   string
	}{
		{
			scenario:        "missing database",
			database:        "other",
			expectedContext: context.Background(),
			expectedError:   `mongo database "other" is not registered to the manager`,
		},
		{
			scenario:        "could not parse filter",
			database:        defaultDatabase,
			filter:          &godog.DocString{Content: `malformed`},
			expectedContext: context.Background(),
			expectedError:   `failed to parse filter: error unmarshaling extjson: invalid JSON input`,
		},
		{
			scenario:        "find error",
			database:        defaultDatabase,
			filter:          &godog.DocString{Content: `{}`},
			result:          []bson.D{{{Key: "ok", Value: 0}}},
			expectedContext: context.Background(),
			expectedError:   `could not find documents in collection "customer": command failed`,
		},
		{
			scenario:        "success without filter",
			database:        defaultDatabase,
			result:          createDocsResponse("db", "customer", docs),
			expectedContext: contextWithDocs(context.Background(), docs),
		},
	}

	for _, tc := range testCases {
		tc := tc

		mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

		mt.Cleanup(func() {
			mt.Close()
		})

		mt.Run(tc.scenario, func(t *mtest.T) {
			t.Parallel()

			t.AddMockResponses(tc.result...)

			m := NewManager(WithDefaultDatabase(t.DB))

			ctx, err := m.searchInCollectionOfDatabase(context.Background(), "customer", tc.database, tc.filter)

			assert.Equal(t, tc.expectedContext, ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestManager_NoDocumentsAreAvailableInCollectionOfDatabase(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		database      string
		result        []bson.D
		expectedError string
	}{
		{
			scenario:      "missing database",
			database:      "other",
			expectedError: `mongo database "other" is not registered to the manager`,
		},
		{
			scenario:      "count error",
			database:      defaultDatabase,
			result:        []bson.D{{{Key: "ok", Value: 0}}},
			expectedError: `could not count documents in collection "customer": command failed`,
		},
		{
			scenario:      "has documents",
			database:      defaultDatabase,
			result:        createCountResponse("db", "customer", 5),
			expectedError: `collection "customer" has 5 document(s), expected none`,
		},
		{
			scenario: "no documents",
			database: defaultDatabase,
			result:   createCountResponse("db", "customer", 0),
		},
	}

	for _, tc := range testCases {
		tc := tc

		mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

		mt.Cleanup(func() {
			mt.Close()
		})

		mt.Run(tc.scenario, func(t *mtest.T) {
			t.Parallel()

			t.AddMockResponses(tc.result...)

			m := NewManager(WithDefaultDatabase(t.DB))

			_, err := m.noDocumentsAreAvailableInCollectionOfDatabase(context.Background(), "customer", tc.database)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestManager_HaveNumberOfDocumentsAvailableInCollectionOfDatabase(t *testing.T) {
	t.Parallel()

	const expectedNumberOfDocuments int64 = 5

	testCases := []struct {
		scenario      string
		database      string
		result        []bson.D
		expectedError string
	}{
		{
			scenario:      "missing database",
			database:      "other",
			expectedError: `mongo database "other" is not registered to the manager`,
		},
		{
			scenario:      "count error",
			database:      defaultDatabase,
			result:        []bson.D{{{Key: "ok", Value: 0}}},
			expectedError: `could not count documents in collection "customer": command failed`,
		},
		{
			scenario:      "mismatched",
			database:      defaultDatabase,
			result:        createCountResponse("db", "customer", 4),
			expectedError: `collection "customer" has 4 document(s), expected 5`,
		},
		{
			scenario: "matched",
			database: defaultDatabase,
			result:   createCountResponse("db", "customer", expectedNumberOfDocuments),
		},
	}

	for _, tc := range testCases {
		tc := tc

		mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

		mt.Cleanup(func() {
			mt.Close()
		})

		mt.Run(tc.scenario, func(t *mtest.T) {
			t.Parallel()

			t.AddMockResponses(tc.result...)

			m := NewManager(WithDefaultDatabase(t.DB))

			_, err := m.haveNumberOfDocumentsAvailableInCollectionOfDatabase(context.Background(), expectedNumberOfDocuments, "customer", tc.database)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestManager_HaveOnlyTheseDocumentsAvailableInCollectionOfDatabase(t *testing.T) {
	t.Parallel()

	fixtures := readFixtures("resources/fixtures/customers.json")
	docs := mustParseDocs(fixtures)

	const expectedIgnoredDiff = `[
		{
			"_id": "<ignored-diff>",
			"name": "John Doe",
			"age": 30,
			"address": {
				"street": "Street 1",
				"city": "City 1",
				"country": "Country 1"
			}
		},
		{
			"_id": "<ignored-diff>",
			"name": "Jane Doe",
			"age": 20,
			"address": {
				"street": "Street 2",
				"city": "City 2",
				"country": "Country 2"
			}
		}
	]`

	testCases := []struct {
		scenario      string
		database      string
		result        []bson.D
		expectedDocs  string
		expectedError string
	}{
		{
			scenario:      "missing database",
			database:      "other",
			expectedError: `mongo database "other" is not registered to the manager`,
		},
		{
			scenario:      "could not parse expected docs",
			database:      defaultDatabase,
			expectedDocs:  `[`,
			expectedError: `failed to parse expected documents: error unmarshaling extjson: invalid JSON input; unexpected end of input at position 0`,
		},
		{
			scenario:      "find error",
			database:      defaultDatabase,
			result:        []bson.D{{{Key: "ok", Value: 0}}},
			expectedDocs:  `[{}]`,
			expectedError: `could not find documents in collection "customer": command failed`,
		},
		{
			scenario:     "mismatched",
			database:     defaultDatabase,
			result:       createCursorResponse("db", "customer", bson.D{{Key: "name", Value: "Jane"}}),
			expectedDocs: `[{"name": "John"}]`,
			expectedError: `not equal:
 [
   {
-    "name": "John"
+    "name": "Jane"
   }
 ]
`,
		},
		{
			scenario:     "matched - equal",
			database:     defaultDatabase,
			result:       createDocsResponse("db", "customer", docs),
			expectedDocs: string(fixtures),
		},
		{
			scenario:     "matched - ignored diff",
			database:     defaultDatabase,
			result:       createDocsResponse("db", "customer", docs),
			expectedDocs: expectedIgnoredDiff,
		},
	}

	for _, tc := range testCases {
		tc := tc

		mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

		mt.Cleanup(func() {
			mt.Close()
		})

		mt.Run(tc.scenario, func(t *mtest.T) {
			t.Parallel()

			t.AddMockResponses(tc.result...)

			m := NewManager(WithDefaultDatabase(t.DB))

			_, err := m.haveOnlyTheseDocumentsAvailableInCollectionOfDatabase(context.Background(), "customer", tc.database, &godog.DocString{Content: tc.expectedDocs})

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestManager_HaveOnlyTheseDocumentsFromFileAvailableInCollectionOfDatabase(t *testing.T) {
	t.Parallel()

	docs := mustParseDocs(readFixtures("resources/fixtures/customers.json"))

	testCases := []struct {
		scenario      string
		database      string
		filePath      string
		result        []bson.D
		expectedError string
	}{
		{
			scenario:      "file not found",
			database:      "other",
			filePath:      "resources/fixtures/unknown.json",
			expectedError: `open resources/fixtures/unknown.json: no such file or directory`,
		},
		{
			scenario:      "missing database",
			database:      "other",
			filePath:      "resources/fixtures/empty.json",
			expectedError: `mongo database "other" is not registered to the manager`,
		},
		{
			scenario:      "could not parse expected docs",
			database:      defaultDatabase,
			filePath:      "resources/fixtures/malformed.json",
			expectedError: `failed to parse expected documents: error unmarshaling extjson: invalid JSON input; unexpected end of input at position 0`,
		},
		{
			scenario:      "find error",
			database:      defaultDatabase,
			filePath:      "resources/fixtures/empty.json",
			result:        []bson.D{{{Key: "ok", Value: 0}}},
			expectedError: `could not find documents in collection "customer": command failed`,
		},
		{
			scenario: "mismatched",
			database: defaultDatabase,
			filePath: "resources/fixtures/empty.json",
			result:   createCursorResponse("db", "customer", bson.D{{Key: "name", Value: "Jane"}}),
			expectedError: `not equal:
 [
+  {
+    "name": "Jane"
+  }
 ]
`,
		},
		{
			scenario: "matched - no documents",
			database: defaultDatabase,
			filePath: "resources/fixtures/empty.json",
			result:   createCursorResponse("db", "customer"),
		},
		{
			scenario: "matched - has documents",
			database: defaultDatabase,
			filePath: "resources/fixtures/customers.json",
			result:   createDocsResponse("db", "customer", docs),
		},
	}

	for _, tc := range testCases {
		tc := tc

		mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

		mt.Cleanup(func() {
			mt.Close()
		})

		mt.Run(tc.scenario, func(t *mtest.T) {
			t.Parallel()

			t.AddMockResponses(tc.result...)

			m := NewManager(WithDefaultDatabase(t.DB))

			_, err := m.haveOnlyTheseDocumentsFromFileAvailableInCollectionOfDatabase(context.Background(), tc.filePath, "customer", tc.database)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestManager_HaveNumberOfDocumentsInSearchResult(t *testing.T) {
	t.Parallel()

	docs := mustParseDocs(readFixtures("resources/fixtures/customers.json"))

	testCases := []struct {
		scenario      string
		context       context.Context // nolint: containedctx
		expected      int64
		expectedError string
	}{
		{
			scenario:      "no docs in context",
			context:       context.Background(),
			expectedError: `no documents are available in the search result, did you forget to search?`,
		},
		{
			scenario: "empty docs in context",
			context:  contextWithDocs(context.Background(), []bsoncore.Document{}),
		},
		{
			scenario:      "mismatched",
			context:       contextWithDocs(context.Background(), docs),
			expected:      1,
			expectedError: `there are 2 documents in the search result, expected 1`,
		},
		{
			scenario: "matched",
			context:  contextWithDocs(context.Background(), docs),
			expected: 2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m := NewManager()

			_, err := m.haveNumberOfDocumentsInSearchResult(tc.context, tc.expected)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestManager_HaveDocumentsInSearchResult(t *testing.T) {
	t.Parallel()

	docs := mustParseDocs([]byte(`[{"name": "John"}]`))

	testCases := []struct {
		scenario      string
		context       context.Context // nolint: containedctx
		expectedDocs  string
		expectedError string
	}{
		{
			scenario:      "no docs in context",
			context:       context.Background(),
			expectedError: `no documents are available in the search result, did you forget to search?`,
		},
		{
			scenario:      "could not parse expected docs",
			context:       contextWithDocs(context.Background(), []bsoncore.Document{}),
			expectedDocs:  `[`,
			expectedError: `failed to parse expected documents: error unmarshaling extjson: invalid JSON input; unexpected end of input at position 0`,
		},
		{
			scenario:     "empty docs in context",
			context:      contextWithDocs(context.Background(), []bsoncore.Document{}),
			expectedDocs: `[]`,
		},
		{
			scenario:     "mismatched",
			context:      contextWithDocs(context.Background(), docs),
			expectedDocs: `[{"name": "Jane"}]`,
			expectedError: `not equal:
 [
   {
-    "name": "Jane"
+    "name": "John"
   }
 ]
`,
		},
		{
			scenario:     "matched",
			context:      contextWithDocs(context.Background(), docs),
			expectedDocs: `[{"name": "John"}]`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m := NewManager()

			_, err := m.haveDocumentsInSearchResult(tc.context, &godog.DocString{Content: tc.expectedDocs})

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func readFixtures(filePath string) []byte { // nolint: unparam
	data, err := os.ReadFile(path.Clean(filePath))
	if err != nil {
		panic(err)
	}

	return data
}

func mustParseDocs(data []byte) []bsoncore.Document {
	docs, err := bytesToDocs(data)
	if err != nil {
		panic(err)
	}

	return docs
}

func mustParseBSOND(data []byte) bson.D {
	docs, err := bytesToBSOND(data)
	if err != nil {
		panic(err)
	}

	return docs
}

func docsToBSOND(docs []bsoncore.Document) []bson.D {
	result := make([]bson.D, len(docs))

	for i, d := range docs {
		result[i] = mustParseBSOND([]byte(d.String()))
	}

	return result
}

func createDocsResponse(dbName, collectionName string, docs []bsoncore.Document) []bson.D { // nolint: unparam
	return createCursorResponse(dbName, collectionName, docsToBSOND(docs)...)
}

func createCountResponse(dbName, collectionName string, count int64) []bson.D { // nolint: unparam
	return createCursorResponse(dbName, collectionName, bson.D{{Key: "n", Value: count}})
}

func createCursorResponse(dbName, collectionName string, batch ...bson.D) []bson.D {
	ns := fmt.Sprintf("%s.%s", dbName, collectionName)

	return []bson.D{
		mtest.CreateCursorResponse(
			1,
			ns,
			mtest.FirstBatch,
			batch...,
		),
		mtest.CreateCursorResponse(
			1,
			ns,
			mtest.NextBatch,
		),
	}
}
