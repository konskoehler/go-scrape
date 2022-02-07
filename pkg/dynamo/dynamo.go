package dynamo

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/konskoehler/go-scrape/pkg/sale"
)

// DB is a DynamoDB service for a particular table.
type DB struct {
	dynamodb *dynamodb.DynamoDB
	table    string
}

// New creates a new DynamoDB session.
func New(region string, table string) (*DB, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	if err != nil {
		return nil, err
	}

	return &DB{
		dynamodb: dynamodb.New(sess),
		table:    table,
	}, nil
}

// puts one sale item into the DynamoDB table.
func (d *DB) PutSale(sale sale.Sale, t time.Time) error {

	av, err := dynamodbattribute.MarshalMap(sale)
	if err != nil {
		return err
	}

	fmt.Printf("marshalled struct: %+v", av)

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: &d.table,
	}

	_, err = d.dynamodb.PutItem(input)

	if err != nil {
		return err
	}

	return nil
}
