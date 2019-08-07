package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/jinzhu/gorm"

	"github.com/bipol/scrapedumper/pkg/postgres"

        //GORM postgres dialect
        _ "github.com/jinzhu/gorm/dialects/postgres"
)

type options struct {
	DataLocation             string `long:"data-location" env:"DATA_LOCATION" description:"local path to from which to collect JSON files" required:"true"`
	PostgresConnectionString string `long:"postgres-connection-string" env:"POSTGRES_CONNECTION_STRING" required:"true"`
}

func main() {
	fmt.Println("Starting postgres loader")
	var opts options
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open("postgres", opts.PostgresConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := postgres.NewRepository(db)
	err = repo.EnsureTables()
	if err != nil {
		log.Fatal(err)
	}

	upserter := postgres.NewUpserter(repo, time.Hour)
	loader := postgres.NewBulkLoader(upserter)
	err = loader.LoadDir(opts.DataLocation)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Success!")
}
