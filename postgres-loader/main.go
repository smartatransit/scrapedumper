package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/spf13/afero"
	"go.uber.org/zap"

	"github.com/bipol/scrapedumper/pkg/bulk"
	"github.com/bipol/scrapedumper/pkg/dumper"
	"github.com/bipol/scrapedumper/pkg/postgres"

	//database/sql driver
	_ "github.com/lib/pq"
)

type options struct {
	DataLocation             string `long:"data-location" env:"DATA_LOCATION" description:"local path to from which to collect JSON files" required:"true"`
	PostgresConnectionString string `long:"postgres-connection-string" env:"POSTGRES_CONNECTION_STRING" required:"true"`
	StartAt                  string `long:"start-at-alphabetically" env:"START_AT_ALPHABETICALLY"`
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

	repo := postgres.NewRepository(db)
	err = repo.EnsureTables()
	if err != nil {
		log.Fatal(err)
	}

	fs := afero.NewOsFs()

	upserter := postgres.NewUpserter(repo, time.Hour)
	dumper := dumper.NewPostgresDumpHandler(logger, upserter)
	dirDumper := bulk.NewDirectoryDumper(fs, dumper)
	err = dirDumper.DumpDirectory(
		context.Background(),
		opts.DataLocation,
		opts.StartAt,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Success!")
}
