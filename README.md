# Cucumber mongodb steps for Golang

[![GitHub Releases](https://img.shields.io/github/v/release/godogx/mongosteps)](https://github.com/godogx/mongosteps/releases/latest)
[![Build Status](https://github.com/godogx/mongosteps/actions/workflows/test.yaml/badge.svg)](https://github.com/godogx/mongosteps/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/godogx/mongosteps/branch/master/graph/badge.svg?token=eTdAgDE2vR)](https://codecov.io/gh/godogx/mongosteps)
[![Go Report Card](https://goreportcard.com/badge/github.com/godogx/mongosteps)](https://goreportcard.com/report/github.com/godogx/mongosteps)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/godogx/mongosteps)

`mongosteps` provides steps for [`cucumber/godog`](https://github.com/cucumber/godog) and makes it easy to run tests with MongoDB.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Usage](#usage)
    - [Setup](#setup)
    - [Notes](#notes)
    - [Steps](#steps)
        - [Delete all documents / Truncate collection](#delete-all-documents--truncate-collection)
        - [Insert documents to collection](#insert-documents-to-collection)
        - [Assert no documents in collection](#assert-no-documents-in-collection)
        - [Assert number of documents in collection](#assert-number-of-documents-in-collection)
        - [Assert all documents in collection](#assert-all-documents-in-collection)
        - [Search for documents](#search-for-documents)

## Prerequisites

- `Go >= 1.17`

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Install

```bash
go get github.com/godogx/mongosteps
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Usage

### Setup

Initiate an `mongosteps.Manager` and register it to the scenario

```go
package mypackage

import (
	"bytes"
	"context"
	"math/rand"
	"testing"

	"github.com/cucumber/godog"
	"github.com/godogx/mongosteps"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestIntegration(t *testing.T) {
	out := bytes.NewBuffer(nil)

	// Create mongodb connection.
	conn, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatalf("could not connect to mongodb: %s", err.Error())
	}

	// Initiate a new manager.
	manager := mongosteps.NewManager(
		mongosteps.WithDefaultDatabase(conn.Database("mydb"), mongosteps.CleanUpAfterScenario("mycollection")),
		// If you have more than 1 database, you can use the following:
		// mongosteps.WithDatabase("other", conn.Database("otherdb"), mongosteps.CleanUpAfterScenario("othercollection")),
	)

	suite := godog.TestSuite{
		Name:                 "Integration",
		TestSuiteInitializer: nil,
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			// Register the client.
			manager.RegisterContext(ctx)
		},
		Options: &godog.Options{
			Strict:    true,
			Output:    out,
			Randomize: rand.Int63(),
		},
	}

	// Run the suite.
	if status := suite.Run(); status != 0 {
		t.Fatal(out.String())
	}
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Notes

- All the JSON (for insertions and assertions) are in [MongoDB Extended JSON (v2)](https://docs.mongodb.com/manual/reference/mongodb-extended-json/) format.
- For assertions, we do support `<ignore-diff>` for any data types.

For example: Given these documents are stored in the collection

```json5
[
    {
        "_id": {"$oid": "6250053966df8910f804c3a7"},
        "name": "John Doe",
        "age": 30,
        "address": {
            "street": "Street 1",
            "city": "City 1",
            "country": "Country 1"
        }
    }
]
```

This assertion matches

```json5
[
    {
        "_id": "<ignore-diff>",
        "name": "John Doe",
        "age": 30,
        "address": "<ignore-diff>"
    }
]
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Steps

#### Delete all documents / Truncate collection

- `no (?:docs|documents) in collection "([^"]*)"$`
- `no (?:docs|documents) in collection "([^"]*)" of database "([^"]*)"$`

For example:

```gherkin
Given no documents in collection "customer"
```

```gherkin
Given no documents in collection "customer" of database "other"
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Insert documents to collection

- `these (?:docs|documents) are(?: stored)? in collection "([^"]*)"[:]?$`
- `(?:docs|documents) from(?: file)? "([^"]*)" are(?: stored)? in collection "([^"]*)"$`
- `these (?:docs|documents) are(?: stored)? in collection "([^"]*)" of database "([^"]*)"[:]?$`
- `(?:docs|documents) from(?: file)? "([^"]*)" are(?: stored)? in collection "([^"]*)" of database "([^"]*)"[:]?$`

For example:

```gherkin
Given these documents are stored in collection "customer":
"""
[
    {
        "_id": {"$oid": "6250053966df8910f804c3a7"},
        "name": "John Doe",
        "age": 30,
        "address": {
            "street": "Street 1",
            "city": "City 1",
            "country": "Country 1"
        }
    }
]
"""
```

```gherkin
Given documents from file "../../resources/fixtures/customers.json" are stored in collection "customer" of database "other"
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Assert no documents in collection

- `no (?:docs|documents) are(?: available)? in collection "([^"]*)"$`
- `no (?:docs|documents) are(?: available)? in collection "([^"]*)" of database "([^"]*)"$`

For example:

```gherkin
Then no documents are available in collection "customer"
```

```gherkin
Then no documents are available in collection "customer" of database "other"
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Assert number of documents in collection

- `collection "([^"]*)" should have ([0-9]+) (?:doc|docs|document|documents)(?: available)?$`
- `there (?:is|are) ([0-9]+) (?:doc|docs|document|documents)(?: available)? in collection "([^"]*)" of database "([^"]*)"$`
- `collection "([^"]*)" of database "([^"]*)" should have ([0-9]+) (?:doc|docs|document|documents)(?: available)?$`
- `there (?:is|are) ([0-9]+) (?:doc|docs|document|documents)(?: available)? in collection "([^"]*)"$`

For example:

```gherkin
Then there are 2 documents in collection "customer"
```

```gherkin
Then collection "customer" of database "other" should have 2 documents
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Assert all documents in collection

Inline documents:

- `collection "([^"]*)" should have only (?:this|these) (?:doc|docs|document|documents)(?: available)?[:]?$`
- `there (?:is|are) only (?:this|these) (?:doc|docs|document|documents)(?: available)? in collection "([^"]*)"[:]?$`
- `collection "([^"]*)" of database "([^"]*)" should have only (?:this|these) (?:doc|docs|document|documents)(?: available)?[:]?$`
- `there (?:is|are) only (?:this|these) (?:doc|docs|document|documents)(?: available)? in collection "([^"]*)" of database "([^"]*)"[:]?$`

From a file:

- `there (?:is|are) only (?:this|these) (?:doc|docs|document|documents) from(?: file)? "([^"]*)"(?: available)? in collection "([^"]*)"[:]?$`
- `there (?:is|are) only (?:this|these) (?:doc|docs|document|documents) from(?: file)? "([^"]*)"(?: available)? in collection "([^"]*)" of database "([^"]*)"[:]?$`

For example:

```gherkin
Then collection "customer" should have only these documents:
"""
[
    {
        "_id": "<ignored-diff>",
        "name": "John Doe",
        "age": 30,
        "address": {
            "street": "Street 1",
            "city": "City 1",
            "country": "Country 1"
        }
    }
]
"""
```

```gherkin
Then there are only these documents from file "../../resources/fixtures/customers.json" in collection "customer" of database "other"
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Search for documents

Without a query:

- `(?:search|find) in collection "([^"]*)"$`
- `(?:search|find) in collection "([^"]*)" of database "([^"]*)"$`

With a query:

- `(?:search|find) in collection "([^"]*)" with query[:]?$`
- `(?:search|find) in collection "([^"]*)" of database "([^"]*)" with query[:]?$`

Assert number of documents found:

- `found ([0-9]+) (?:doc|docs|document|documents) in the result$`
- `there (?:is|are) ([0-9]+) (?:doc|docs|document|documents) in the result$`

Assert documents found:

- `(?:this|these) (?:doc|docs|document|documents) (?:is|are) in the result[:]?$`
- `found (?:this|these) (?:doc|docs|document|documents) in the result[:]?$`

For example:

```gherkin
Given documents from file "../../resources/fixtures/customers.json" are stored in collection "customer"

When I search in collection "customer"

Then I found 2 documents in the result
And I found these documents in the result:
"""
[
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
]
"""
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)
