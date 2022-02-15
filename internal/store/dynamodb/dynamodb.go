package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
	db        *dynamodb.Client
	table     string
	partition string
}

// New creates a new store, backed by the provided DynamoDB client, that writes
// groups into the provided table using the provided partition key. See the
// Store documentation for details.
func New(db *dynamodb.Client, table, partition string) (Store, error) {
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
func (s Store) List(ctx context.Context) ([]string, error) {
	expr, err := expression.NewBuilder().
		WithKeyCondition(
			expression.KeyEqual(
				expression.Key(partitionKey), expression.Value(s.partition),
			),
		).
		WithProjection(expression.NamesList(
			expression.Name(groupKey),
		)).
		Build()
	if err != nil {
		return nil, errors.Wrap(err, "building expression")
	}

	result, err := s.db.Query(ctx, &dynamodb.QueryInput{
		TableName:                 &s.table,
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "listing groups for %q from table %q", s.partition, s.table)
	}

	list := make([]string, len(result.Items))
	for i, item := range result.Items {
		v, ok := item[groupKey].(*types.AttributeValueMemberS)
		if !ok {
			return nil, errors.Errorf("invalid type %T in group names", item[groupKey])
		}
		list[i] = v.Value
	}
	return list, nil
}

// Get obtains the options in a single named group from this Store's partition.
func (s Store) Get(ctx context.Context, name string) ([]string, error) {
	expr, err := expression.NewBuilder().
		WithProjection(expression.NamesList(
			expression.Name(itemsKey),
		)).
		Build()
	if err != nil {
		return nil, errors.Wrap(err, "building expression")
	}

	result, err := s.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &s.table,
		Key: map[string]types.AttributeValue{
			partitionKey: &types.AttributeValueMemberS{Value: s.partition},
			groupKey:     &types.AttributeValueMemberS{Value: name},
		},
		ProjectionExpression:     expr.Projection(),
		ExpressionAttributeNames: expr.Names(),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "getting %q for %q from table %q", name, s.partition, s.table)
	}

	if len(result.Item) == 0 {
		return nil, nil
	}

	v, ok := result.Item[itemsKey].(*types.AttributeValueMemberSS)
	if !ok {
		return nil, errors.Errorf("invalid type %T in group items", v)
	}
	return v.Value, nil
}

// Put saves the provided options into a named group for this Store's
// partition.
func (s Store) Put(ctx context.Context, name string, options []string) error {
	_, err := s.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &s.table,
		Item: map[string]types.AttributeValue{
			partitionKey: &types.AttributeValueMemberS{Value: s.partition},
			groupKey:     &types.AttributeValueMemberS{Value: name},
			itemsKey:     &types.AttributeValueMemberSS{Value: options},
		},
	})
	return errors.Wrapf(err, "saving %q for %q to table %q", name, s.partition, s.table)
}

// Delete removes the named group from this Store's partition.
func (s Store) Delete(ctx context.Context, name string) (bool, error) {
	result, err := s.db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &s.table,
		Key: map[string]types.AttributeValue{
			partitionKey: &types.AttributeValueMemberS{Value: s.partition},
			groupKey:     &types.AttributeValueMemberS{Value: name},
		},
		ReturnValues: types.ReturnValueAllOld,
	})
	existed := len(result.Attributes) > 0
	return existed, errors.Wrapf(err, "deleting %q for %q from table %q", name, s.partition, s.table)
}
