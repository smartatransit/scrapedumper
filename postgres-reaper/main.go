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
	DataLocation             string `long:"data-location" env:"DATA_LOCATION" description:"local path to from which to collect JSON files" required:"true"`
	PostgresConnectionString string `long:"postgres-connection-string" env:"POSTGRES_CONNECTION_STRING" required:"true"`
	RunTTLSeconds            int    `long:"run-ttl-seconds" env:"RUN_TTL_SECONDS"`
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

	threshold := time.Now().Add(-time.Second * time.Duration(opts.RunTTLSeconds))
	if err := repo.DeleteStaleRuns(postgres.EasternTime(threshold)); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Success!")
}
