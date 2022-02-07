package dynamo

import (
	"log"
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

func (d *DB) PutSales(sales []sale.Sale, t time.Time) error {

	for _, item := range sales {
		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			log.Fatalf("Got error marshalling map: %s", err)
			return err
		}

		// Create item in table Movies
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: &d.table,
		}

		_, err = d.dynamodb.PutItem(input)
		if err != nil {
			log.Fatalf("Got error calling PutItem: %s", err)
			return err
		}
	}
	return nil
}
