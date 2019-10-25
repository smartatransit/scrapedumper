package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"

	"github.com/bipol/scrapedumper/pkg/postgres"

	//database/sql driver
	_ "github.com/lib/pq"
)

type options struct {
	PostgresConnectionString string    `long:"postgres-connection-string" env:"POSTGRES_CONNECTION_STRING" required:"true"`
	RunThreshold             time.Time `long:"run-threshold-rfc3339" env:"RUN_THRESHOLD_RFC3339" description:"The moment before which we should delete any runs. Formatted as an RFC3339 timestamp."`
}

func main() {
	fmt.Println("Starting postgres loader")
	var opts options
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync() // flushes buffer, if any
	}()

	db, err := sql.Open("postgres", opts.PostgresConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := postgres.NewRepository(logger, db)
	err = repo.EnsureTables()
	if err != nil {
		log.Fatal(err)
	}

	if err := repo.DeleteStaleRuns(postgres.EasternTime(opts.RunThreshold)); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Success!")
}
