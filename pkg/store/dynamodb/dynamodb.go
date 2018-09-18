package dynamodb // import "go.alexhamlin.co/pkg/randomizer/store/dynamodb"

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pkg/errors"
)

// DefaultTable is the default name of the DynamoDB table used by Stores.
const DefaultTable = "RandomizerGroups"

// DefaultPartition is the default value of the partition key used by Stores.
const DefaultPartition = "Groups"

const (
	partitionKey = "Partition"
	groupKey     = "Group"
)

// Store is a store backed by a pre-existing Amazon DynamoDB table.
//
// The DynamoDB table used by a Store must have a composite primary key, with a
// partition key named "Partition" and a sort key named "Group", both
// string-valued. Items in each row are stored in a string set column named
// "Items".
type Store struct {
	db        *dynamodb.DynamoDB
	table     string
	partition string
}

// Option represents a type for options that can be applied to a Store.
type Option func(*Store)

// New creates a new store, backed by the provided DynamoDB client, that writes
// groups using the provided partition key. See the Store documentation for
// details.
func New(db *dynamodb.DynamoDB, options ...Option) Store {
	store := Store{
		db:        db,
		table:     DefaultTable,
		partition: DefaultPartition,
	}

	for _, opt := range options {
		opt(&store)
	}

	return store
}

// WithTable creates a Store that uses the DynamoDB table with the provided
// name, rather than DefaultTable.
func WithTable(table string) Option {
	return func(s *Store) {
		s.table = table
	}
}

// WithPartition creates a Store that uses the provided value for the partition
// key in the DynamoDB table.
func WithPartition(partition string) Option {
	return func(s *Store) {
		s.partition = partition
	}
}

// Provision requests the creation of a DynamoDB table, using this store's
// table name and the expected schema for a Store, and the provided numbers of
// read and write capacity units.
//
// Note that Provision returns when DynamoDB has accepted the request to create
// a table, not necessarily when the table is ready to accept reads or writes.
// Provision will fail if the table already exists. Generally speaking,
// Provision is a helper method for use outside of a normal application
// workflow.
func (s Store) Provision(readCap, writeCap int64) error {
	var (
		partitionKey = partitionKey
		groupKey     = groupKey
	)

	input := dynamodb.CreateTableInput{
		TableName: &s.table,

		KeySchema: []dynamodb.KeySchemaElement{
			{
				AttributeName: &partitionKey,
				KeyType:       dynamodb.KeyTypeHash,
			},
			{
				AttributeName: &groupKey,
				KeyType:       dynamodb.KeyTypeRange,
			},
		},

		AttributeDefinitions: []dynamodb.AttributeDefinition{
			{
				AttributeName: &partitionKey,
				AttributeType: dynamodb.ScalarAttributeTypeS, // String
			},
			{
				AttributeName: &groupKey,
				AttributeType: dynamodb.ScalarAttributeTypeS, // String
			},
		},

		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  &readCap,
			WriteCapacityUnits: &writeCap,
		},
	}

	_, err := s.db.CreateTableRequest(&input).Send()
	return errors.Wrap(err, "creating DynamoDB table")
}
