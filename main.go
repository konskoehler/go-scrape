package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/konskoehler/go-scrape/pkg/collector"
	"github.com/konskoehler/go-scrape/pkg/dynamo"
	"github.com/konskoehler/go-scrape/pkg/sale"
)

// declare global variables to be used
var threads, queueStorage int
var baseUrl = "https://www.ebay.de/sch/i.html?_from=R40&_sacat=0&LH_Sold=1&_udlo=&_udhi=&_samilow=&_samihi=&_sadis=15&_stpos=10437&_sop=12&_dmd=1&LH_Complete=1&_fosrp=1&_dcat=183454&Bewertet=Ja&_nkw=pokemon+holo&_ipg=240&rt=nc"

// HandleRequest handles one request to the lambda function.
func HandleRequest(ctx context.Context) error {

	tablename := os.Getenv("DYNAMODB_TABLE")
	region := os.Getenv("DYNAMODB_REGION")

	db, err := dynamo.New(region, tablename)

	if err != nil {
		log.Fatal(err)
	}

	threads = 4
	queueStorage = 10000
	var sales []sale.Sale

	collector.Run(threads, queueStorage, baseUrl, &sales)

	for _, s := range sales {
		err := db.PutSale(s)
		if err != nil {
			log.Print(err)
		}
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
