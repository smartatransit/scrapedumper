package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bipol/scrapedumper/pkg/postgres"

	"github.com/jessevdk/go-flags"
	"github.com/jinzhu/gorm"

	//GORM postgres dialect
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type options struct {
	OutputLocation  string `long:"output-location" env:"OUTPUT_LOCATION" description:"local path to output"`
	DynamoTableName string `long:"dynamo-table-name" env:"DYNAMO_TABLE_NAME" description:"dynamo table name"`
	S3BucketName    string `long:"s3-bucket-name" env:"S3_BUCKET_NAME" description:"s3 bucket to dump stuff into"`
	// MartaAPIKey       string `long:"marta-api-key" env:"MARTA_API_KEY" description:"marta api key" required:"true"`
	// PollTimeInSeconds int    `long:"poll-time-in-seconds" env:"POLL_TIME_IN_SECONDS" description:"time to poll marta api every second" required:"true"`
}

func main() {
	fmt.Println("Starting scrape and dump")
	var opts options
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open("postgres", "host=localhost user=newuser dbname=smartatransit sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.LogMode(false)

	repo := postgres.RepositoryAgent{DB: db}
	if err = repo.EnsureTables(); err != nil {
		log.Fatal(err)
	}

	upserter := postgres.UpserterAgent{}

	if err = repo.EnsureArrivalRecord("N", "GOLD", "my train", time.Date(2000, 1, 1, 5, 34, 0, 0, time.Local), "FIVE POINTS STATION"); err != nil {
		log.Fatal(err)
	}

	if err = repo.AddArrivalEstimate(
		"N", "GOLD", "my train", time.Date(2000, 1, 1, 5, 34, 0, 0, time.Local), "FIVE POINTS STATION",
		postgres.ArrivalEstimate{
			EventTime:            time.Date(2000, 1, 1, 5, 34, 1, 0, time.Local),
			EstimatedArrivalTime: time.Date(2001, 1, 1, 5, 34, 1, 0, time.Local),
		},
	); err != nil {
		log.Fatal(err)
	}

	if err = repo.AddArrivalEstimate(
		"N", "GOLD", "my train", time.Date(2000, 1, 1, 5, 34, 0, 0, time.Local), "FIVE POINTS STATION",
		postgres.ArrivalEstimate{
			EventTime:            time.Date(2000, 1, 1, 5, 34, 2, 0, time.Local),
			EstimatedArrivalTime: time.Date(2001, 1, 1, 5, 34, 2, 0, time.Local),
		},
	); err != nil {
		log.Fatal(err)
	}

	if err = repo.SetArrivalTime(
		"N", "GOLD", "my train", time.Date(2000, 1, 1, 5, 34, 0, 0, time.Local), "FIVE POINTS STATION",
		time.Date(2003, 1, 1, 5, 34, 2, 0, time.Local),
	); err != nil {
		log.Fatal(err)
	}

	fmt.Println("success")
}
