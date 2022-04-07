package mongosteps

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/cucumber/godog"
	"github.com/swaggest/assertjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultDatabase = "default"

// ManagerOption sets an option on the Manager.
type ManagerOption interface {
	applyManagerOption(*Manager)
}

type managerOptionFunc func(*Manager)

func (f managerOptionFunc) applyManagerOption(m *Manager) {
	f(m)
}

// Manager manages all databases for running cucumber steps.
type Manager struct {
	databases map[string]*database
}

// RegisterContext registers the manager to godog scenarios.
func (m *Manager) RegisterContext(sc *godog.ScenarioContext) {
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		for _, db := range m.databases {
			if err := db.cleanUp(ctx); err != nil {
				return ctx, err
			}
		}

		return ctx, nil
	})

	m.registerSteps(sc)
	m.registerAssertions(sc)
}

// registerSteps registers the godog setup steps.
func (m *Manager) registerSteps(sc *godog.ScenarioContext) {
	sc.Step(`no (?:docs|documents) in collection "([^"]*)"$`,
		func(ctx context.Context, collectionName string) (context.Context, error) {
			return m.noDocumentsInCollectionOfDatabase(ctx, collectionName, defaultDatabase)
		},
	)

	sc.Step(`these (?:docs|documents) are(?: stored)? in collection "([^"]*)"[:]?$`,
		func(ctx context.Context, collectionName string, data *godog.DocString) (context.Context, error) {
			return m.theseDocumentsAreStoredInCollectionOfDatabase(ctx, collectionName, defaultDatabase, data)
		},
	)

	sc.Step(`(?:docs|documents) from(?: file)? "([^"]*)" are(?: stored)? in collection "([^"]*)"$`,
		func(ctx context.Context, filePath, collectionName string) (context.Context, error) {
			return m.theseDocumentsFromFileAreStoredInCollectionOfDatabase(ctx, filePath, collectionName, defaultDatabase)
		},
	)

	sc.Step(`(?:search|find) in collection "([^"]*)"$`,
		func(ctx context.Context, collectionName string) (context.Context, error) {
			return m.searchInCollectionOfDatabase(ctx, collectionName, defaultDatabase, nil)
		},
	)

	sc.Step(`(?:search|find) in collection "([^"]*)" with query[:]?$`,
		func(ctx context.Context, collectionName string, data *godog.DocString) (context.Context, error) {
			return m.searchInCollectionOfDatabase(ctx, collectionName, defaultDatabase, data)
		},
	)

	sc.Step(`no (?:docs|documents) in collection "([^"]*)" of database "([^"]*)"$`, m.noDocumentsInCollectionOfDatabase)
	sc.Step(`these (?:docs|documents) are(?: stored)? in collection "([^"]*)" of database "([^"]*)"[:]?$`, m.theseDocumentsAreStoredInCollectionOfDatabase)
	sc.Step(`(?:docs|documents) from(?: file)? "([^"]*)" are(?: stored)? in collection "([^"]*)" of database "([^"]*)"[:]?$`, m.theseDocumentsFromFileAreStoredInCollectionOfDatabase)
	sc.Step(`(?:search|find) in collection "([^"]*)" of database "([^"]*)" with query[:]?$`, m.searchInCollectionOfDatabase)

	sc.Step(`(?:search|find) in collection "([^"]*)" of database "([^"]*)"$`,
		func(ctx context.Context, collectionName, databaseName string) (context.Context, error) {
			return m.searchInCollectionOfDatabase(ctx, collectionName, databaseName, nil)
		},
	)
}

// registerAssertions registers the godog assertions steps.
func (m *Manager) registerAssertions(sc *godog.ScenarioContext) {
	sc.Step(`no (?:docs|documents) are(?: available)? in collection "([^"]*)"$`,
		func(ctx context.Context, collectionName string) (context.Context, error) {
			return m.noDocumentsAreAvailableInCollectionOfDatabase(ctx, collectionName, defaultDatabase)
		},
	)

	sc.Step(`there (?:is|are) ([0-9]+) (?:doc|docs|document|documents)(?: available)? in collection "([^"]*)"$`,
		func(ctx context.Context, count int64, collectionName string) (context.Context, error) {
			return m.haveNumberOfDocumentsAvailableInCollectionOfDatabase(ctx, count, collectionName, defaultDatabase)
		},
	)

	sc.Step(`collection "([^"]*)" should have ([0-9]+) (?:doc|docs|document|documents)(?: available)?$`,
		func(ctx context.Context, collectionName string, count int64) (context.Context, error) {
			return m.haveNumberOfDocumentsAvailableInCollectionOfDatabase(ctx, count, collectionName, defaultDatabase)
		},
	)

	sc.Step(`there (?:is|are) only (?:this|these) (?:doc|docs|document|documents)(?: available)? in collection "([^"]*)"[:]?$`,
		func(ctx context.Context, collectionName string, data *godog.DocString) (context.Context, error) {
			return m.haveOnlyTheseDocumentsAvailableInCollectionOfDatabase(ctx, collectionName, defaultDatabase, data)
		},
	)

	sc.Step(`collection "([^"]*)" should have only (?:this|these) (?:doc|docs|document|documents)(?: available)?[:]?$`,
		func(ctx context.Context, collectionName string, data *godog.DocString) (context.Context, error) {
			return m.haveOnlyTheseDocumentsAvailableInCollectionOfDatabase(ctx, collectionName, defaultDatabase, data)
		},
	)

	sc.Step(`there (?:is|are) only (?:this|these) (?:doc|docs|document|documents) from(?: file)? "([^"]*)"(?: available)? in collection "([^"]*)"[:]?$`,
		func(ctx context.Context, filePath, collectionName string) (context.Context, error) {
			return m.haveOnlyTheseDocumentsFromFileAvailableInCollectionOfDatabase(ctx, filePath, collectionName, defaultDatabase)
		},
	)

	sc.Step(`no (?:docs|documents) are(?: available)? in collection "([^"]*)" of database "([^"]*)"$`, m.noDocumentsAreAvailableInCollectionOfDatabase)
	sc.Step(`there (?:is|are) ([0-9]+) (?:doc|docs|document|documents)(?: available)? in collection "([^"]*)" of database "([^"]*)"$`, m.haveNumberOfDocumentsAvailableInCollectionOfDatabase)
	sc.Step(`there (?:is|are) only (?:this|these) (?:doc|docs|document|documents)(?: available)? in collection "([^"]*)" of database "([^"]*)"[:]?$`, m.haveOnlyTheseDocumentsAvailableInCollectionOfDatabase)
	sc.Step(`collection "([^"]*)" of database "([^"]*)" should have only (?:this|these) (?:doc|docs|document|documents)(?: available)?[:]?$`, m.haveOnlyTheseDocumentsAvailableInCollectionOfDatabase)
	sc.Step(`there (?:is|are) only (?:this|these) (?:doc|docs|document|documents) from(?: file)? "([^"]*)"(?: available)? in collection "([^"]*)" of database "([^"]*)"[:]?$`, m.haveOnlyTheseDocumentsFromFileAvailableInCollectionOfDatabase)

	sc.Step(`collection "([^"]*)" of database "([^"]*)" should have ([0-9]+) (?:doc|docs|document|documents)(?: available)?$`,
		func(ctx context.Context, collectionName, databaseName string, count int64) (context.Context, error) {
			return m.haveNumberOfDocumentsAvailableInCollectionOfDatabase(ctx, count, collectionName, databaseName)
		},
	)

	sc.Step(`found ([0-9]+) (?:doc|docs|document|documents) in the result$`, m.haveNumberOfDocumentsInSearchResult)
	sc.Step(`there (?:is|are) ([0-9]+) (?:doc|docs|document|documents) in the result$`, m.haveNumberOfDocumentsInSearchResult)
	sc.Step(`found (?:this|these) (?:doc|docs|document|documents) in the result[:]?$`, m.haveDocumentsInSearchResult)
	sc.Step(`(?:this|these) (?:doc|docs|document|documents) (?:is|are) in the result[:]?$`, m.haveDocumentsInSearchResult)
}

func (m *Manager) getDatabase(dbName string) (*database, error) {
	db, ok := m.databases[dbName]
	if !ok {
		return nil, fmt.Errorf("mongo database %q is not registered to the manager", dbName) // nolint: goerr113
	}

	return db, nil
}

func (m *Manager) noDocumentsInCollectionOfDatabase(ctx context.Context, collectionName string, dbName string) (context.Context, error) {
	db, err := m.getDatabase(dbName)
	if err != nil {
		return ctx, err
	}

	return ctx, db.truncate(ctx, collectionName)
}

func (m *Manager) theseDocumentsAreStoredInCollectionOfDatabase(ctx context.Context, collectionName, dbName string, data *godog.DocString) (context.Context, error) {
	db, err := m.getDatabase(dbName)
	if err != nil {
		return ctx, err
	}

	docs, err := stringToDocs(data)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse documents: %w", err)
	}

	return ctx, db.store(ctx, collectionName, docs)
}

func (m *Manager) theseDocumentsFromFileAreStoredInCollectionOfDatabase(ctx context.Context, filePath, collectionName, dbName string) (context.Context, error) {
	db, err := m.getDatabase(dbName)
	if err != nil {
		return ctx, err
	}

	data, err := os.ReadFile(path.Clean(filePath))
	if err != nil {
		return ctx, err
	}

	docs, err := bytesToDocs(data)
	if err != nil {
		return ctx, err
	}

	return ctx, db.store(ctx, collectionName, docs)
}

func (m *Manager) searchInCollectionOfDatabase(ctx context.Context, collectionName string, dbName string, data *godog.DocString) (context.Context, error) {
	db, err := m.getDatabase(dbName)
	if err != nil {
		return ctx, err
	}

	filter, err := stringToBSOND(data)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse filter: %w", err)
	}

	result, err := db.find(ctx, collectionName, filter, options.Find().SetLimit(0).SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return ctx, err
	}

	return contextWithDocs(ctx, result), nil
}

func (m *Manager) noDocumentsAreAvailableInCollectionOfDatabase(ctx context.Context, collectionName string, dbName string) (context.Context, error) {
	db, err := m.getDatabase(dbName)
	if err != nil {
		return ctx, err
	}

	count, err := db.count(ctx, collectionName, bson.D{})
	if err != nil {
		return ctx, err
	}

	if count > 0 {
		return ctx, fmt.Errorf("collection %q has %d document(s), expected none", collectionName, count) // nolint: goerr113
	}

	return ctx, nil
}

func (m *Manager) haveNumberOfDocumentsAvailableInCollectionOfDatabase(ctx context.Context, expected int64, collectionName string, dbName string) (context.Context, error) {
	db, err := m.getDatabase(dbName)
	if err != nil {
		return ctx, err
	}

	actual, err := db.count(ctx, collectionName, bson.D{})
	if err != nil {
		return ctx, err
	}

	if actual != expected {
		return ctx, fmt.Errorf("collection %q has %d document(s), expected %d", collectionName, actual, expected) // nolint: goerr113
	}

	return ctx, nil
}

func (m *Manager) haveOnlyTheseDocumentsAvailableInCollectionOfDatabase(ctx context.Context, collectionName string, dbName string, data *godog.DocString) (context.Context, error) {
	db, err := m.getDatabase(dbName)
	if err != nil {
		return ctx, err
	}

	expectedDocs, err := stringToDocs(data)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse expected documents: %w", err)
	}

	actualDocs, err := db.find(ctx, collectionName, bson.D{}, options.Find().SetLimit(0).SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return ctx, err
	}

	expected, err := docsToExtJSON(expectedDocs)
	if err != nil {
		return ctx, fmt.Errorf("failed to convert expected documents to JSON: %w", err)
	}

	actual, err := docsToExtJSON(actualDocs)
	if err != nil {
		return ctx, fmt.Errorf("failed to convert actual documents to JSON: %w", err)
	}

	return ctx, assertjson.FailNotEqual(expected, actual)
}

func (m *Manager) haveOnlyTheseDocumentsFromFileAvailableInCollectionOfDatabase(ctx context.Context, filePath string, collectionName string, dbName string) (context.Context, error) {
	expected, err := os.ReadFile(path.Clean(filePath))
	if err != nil {
		return ctx, err
	}

	return m.haveOnlyTheseDocumentsAvailableInCollectionOfDatabase(ctx, collectionName, dbName, &godog.DocString{Content: string(expected)})
}

func (m *Manager) haveNumberOfDocumentsInSearchResult(ctx context.Context, expected int64) (context.Context, error) {
	docs := docsFromContext(ctx)
	if docs == nil {
		//goland:noinspection GoErrorStringFormat
		return ctx, fmt.Errorf("no documents are available in the search result, did you forget to search?") // nolint: goerr113
	}

	actual := int64(len(docs))

	if actual != expected {
		return ctx, fmt.Errorf("there are %d documents in the search result, expected %d", actual, expected) // nolint: goerr113
	}

	return ctx, nil
}

func (m *Manager) haveDocumentsInSearchResult(ctx context.Context, data *godog.DocString) (context.Context, error) {
	actualDocs := docsFromContext(ctx)
	if actualDocs == nil {
		//goland:noinspection GoErrorStringFormat
		return ctx, fmt.Errorf("no documents are available in the search result, did you forget to search?") // nolint: goerr113
	}

	expectedDocs, err := stringToDocs(data)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse expected documents: %w", err)
	}

	expected, err := docsToExtJSON(expectedDocs)
	if err != nil {
		return ctx, fmt.Errorf("failed to convert expected documents to JSON: %w", err)
	}

	actual, err := docsToExtJSON(actualDocs)
	if err != nil {
		return ctx, fmt.Errorf("failed to convert actual documents to JSON: %w", err)
	}

	return ctx, assertjson.FailNotEqual(expected, actual)
}

// NewManager creates a new Manager.
func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{
		databases: make(map[string]*database),
	}

	for _, opt := range opts {
		opt.applyManagerOption(m)
	}

	return m
}

// WithDefaultDatabase sets the default database of the manager.
func WithDefaultDatabase(db *mongo.Database, opts ...DatabaseOption) ManagerOption {
	return managerOptionFunc(func(m *Manager) {
		m.databases[defaultDatabase] = newDatabase(db, opts...)
	})
}

// WithDatabase adds a database to the manager.
func WithDatabase(name string, db *mongo.Database, opts ...DatabaseOption) ManagerOption {
	return managerOptionFunc(func(m *Manager) {
		m.databases[name] = newDatabase(db, opts...)
	})
}
