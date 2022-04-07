Feature: Success cases with all the steps of a specific database.

    Scenario: Collection is empty then no documents should be returned.
        Given no documents in collection "customer" of database "other"

        Then no documents are available in collection "customer" of database "other"
        And there is 0 document in collection "customer" of database "other"
        And collection "customer" of database "other" should have 0 document

    Scenario: Collection is not empty then truncated.
        Given these documents are stored in collection "customer" of database "other":
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
            },
            {
                "_id": {"$oid": "6250053966df8910f804c3a8"},
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

        Then there are 2 documents in collection "customer" of database "other"
        And collection "customer" of database "other" should have 2 documents

        When no documents in collection "customer" of database "other"

        Then no documents are available in collection "customer" of database "other"
        And there is 0 document in collection "customer" of database "other"
        And collection "customer" of database "other" should have 0 document

    Scenario: Store documents from file.
        Given documents from file "../../resources/fixtures/customers.json" are stored in collection "customer" of database "other"

        Then there are 2 documents in collection "customer" of database "other"
        And collection "customer" of database "other" should have 2 documents

        When no documents in collection "customer" of database "other"

        Then no documents are available in collection "customer" of database "other"
        And there is 0 document in collection "customer" of database "other"
        And collection "customer" of database "other" should have 0 document

    Scenario: Collection should have only the stored documents.
        Given documents from file "../../resources/fixtures/customers.json" are stored in collection "customer" of database "other"

        Then there are 2 documents in collection "customer" of database "other"
        And there are only these documents in collection "customer" of database "other":
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
        And collection "customer" of database "other" should have only these documents:
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
        And there are only these documents from file "../../resources/fixtures/customers.json" in collection "customer" of database "other"

    Scenario: Search in collection without a query
        Given documents from file "../../resources/fixtures/customers.json" are stored in collection "customer" of database "other"

        When I search in collection "customer" of database "other"

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

        And there are 2 documents in the result
        And these documents are in the result:
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

    Scenario: Search in collection with a query
        Given documents from file "../../resources/fixtures/customers.json" are stored in collection "customer" of database "other"

        When I search in collection "customer" of database "other" with query:
        """
        {
            "age": {
                "$gt": 25
            }
        }
        """

        Then I found 1 documents in the result
        And I found this document in the result:
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

        And there is 1 documents in the result
        And this document is in the result:
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
