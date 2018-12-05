package dynamodb // import "go.alexhamlin.co/randomizer/internal/store/dynamodb"

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pkg/errors"
)

const (
	partitionKey = "Partition"
	groupKey     = "Group"
	itemsKey     = "Items"
)

// Store is a store backed by a pre-existing Amazon DynamoDB table.
//
// The DynamoDB table used by a Store must have a composite primary key, with a
// partition key named "Partition" and a sort key named "Group", both
// string-valued. Items in each row are stored in a string set attribute named
// "Items".
type Store struct {
	db        *dynamodb.DynamoDB
	table     string
	partition string
}

// New creates a new store, backed by the provided DynamoDB client, that writes
// groups into the provided table using the provided partition key. See the
// Store documentation for details.
func New(db *dynamodb.DynamoDB, table, partition string) (Store, error) {
	if db == nil {
		return Store{}, errors.New("DynamoDB instance is required")
	}

	if table == "" {
		return Store{}, errors.New("table is required")
	}

	if partition == "" {
		return Store{}, errors.New("partition is required")
	}

	return Store{
		db:        db,
		table:     table,
		partition: partition,
	}, nil
}

// List obtains the list of stored groups for this Store's partition.
func (s Store) List() ([]string, error) {
	// Look up rows where the "Partition" attribute is equal to this Store's
	// partition. Note that an expression attribute name is required, as
	// "partition" is a reserved word in DynamoDB.
	keyConditionExpr := "#PART = :value"

	// "Group" is also a reserved word.
	projectionExpr := "#GROUP"

	input := dynamodb.QueryInput{
		TableName:              &s.table,
		KeyConditionExpression: &keyConditionExpr,
		ProjectionExpression:   &projectionExpr,
		ExpressionAttributeNames: map[string]string{
			"#PART":  partitionKey,
			"#GROUP": groupKey,
		},
		ExpressionAttributeValues: map[string]dynamodb.AttributeValue{
			":value": {S: &s.partition},
		},
	}

	result, err := s.db.QueryRequest(&input).Send()
	if err != nil {
		return nil, errors.Wrapf(err, "listing groups for %q from table %q", s.partition, s.table)
	}

	list := make([]string, len(result.Items))
	for i, item := range result.Items {
		list[i] = *item[groupKey].S
	}
	return list, nil
}

// Get obtains the options in a single named group from this Store's partition.
func (s Store) Get(name string) ([]string, error) {
	// "Items" is a reserved word in DynamoDB.
	projectionExpr := "#ITEMS"

	input := dynamodb.GetItemInput{
		TableName: &s.table,
		Key: map[string]dynamodb.AttributeValue{
			partitionKey: {S: &s.partition},
			groupKey:     {S: &name},
		},
		ProjectionExpression:     &projectionExpr,
		ExpressionAttributeNames: map[string]string{"#ITEMS": itemsKey},
	}

	result, err := s.db.GetItemRequest(&input).Send()
	if err != nil {
		return nil, errors.Wrapf(err, "getting %q for %q from table %q", name, s.partition, s.table)
	}

	if len(result.Item) == 0 {
		return nil, nil
	}

	return result.Item[itemsKey].SS, nil
}

// Put saves the provided options into a named group for this Store's
// partition.
func (s Store) Put(name string, options []string) error {
	input := dynamodb.PutItemInput{
		TableName: &s.table,
		Item: map[string]dynamodb.AttributeValue{
			partitionKey: {S: &s.partition},
			groupKey:     {S: &name},
			itemsKey:     {SS: options},
		},
	}

	_, err := s.db.PutItemRequest(&input).Send()
	return errors.Wrapf(err, "saving %q for %q to table %q", name, s.partition, s.table)
}

// Delete removes the named group from this Store's partition.
func (s Store) Delete(name string) (bool, error) {
	input := dynamodb.DeleteItemInput{
		TableName: &s.table,
		Key: map[string]dynamodb.AttributeValue{
			partitionKey: {S: &s.partition},
			groupKey:     {S: &name},
		},
		ReturnValues: dynamodb.ReturnValueAllOld,
	}

	result, err := s.db.DeleteItemRequest(&input).Send()
	existed := len(result.Attributes) > 0
	return existed, errors.Wrapf(err, "deleting %q for %q from table %q", name, s.partition, s.table)
}
