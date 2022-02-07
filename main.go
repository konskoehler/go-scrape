package main

import (
	"github.com/konskoehler/go-scrape/pkg/collector"
	"github.com/konskoehler/go-scrape/pkg/sale"
)

// declare global variables to be used
var threads, queueStorage int
var baseUrl = "https://www.ebay.de/sch/i.html?_from=R40&_sacat=0&LH_Sold=1&_udlo&_udhi&_samilow&_samihi&_sadis=15&_stpos=10437&_sop=12&_dmd=1&_ipg=50&LH_Complete=1&_fosrp=1&_nkw=pokemon%20holo&_dcat=183454&Bewertet=Ja&rt=nc&_trksid=p2045573.m1684"

func main() {

	threads = 4
	queueStorage = 10000
	var sales []sale.Sale

	collector.Run(threads, queueStorage, baseUrl, &sales)

}

// HandleRequest handles one request to the lambda function.
/* func HandleRequest() {
	tablename := os.Getenv("DYNAMODB_TABLE")
	region := os.Getenv("DYNAMODB_REGION")

	db, err := dynamo.New(region, tablename)

	if err != nil {
		log.Fatal(err)
	}


	t := time.Now()



}
*/
