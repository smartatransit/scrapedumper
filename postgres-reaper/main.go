package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"

	"github.com/smartatransit/scrapedumper/pkg/postgres"

	//database/sql driver
	_ "github.com/lib/pq"
)

type options struct {
	PostgresConnectionString string `long:"postgres-connection-string" env:"POSTGRES_CONNECTION_STRING" required:"true"`
	RunTTLMinutes            int    `long:"run-ttl-minues" env:"RUN_TTL_MINUTES3339" description:"The TTL of a run in minues."`
}

func main() {
	fmt.Println("Starting postgres run reaper")
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

	threshold := time.Now().Add(-time.Minute * time.Duration(opts.RunTTLMinutes))
	estimatesDropped, arrivalsDropped, runsDropped, err := repo.DeleteStaleRuns(postgres.EasternTime(threshold))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Success:")
	fmt.Println("Estimates dropped:", estimatesDropped)
	fmt.Println("Arrivals dropped:", arrivalsDropped)
	fmt.Println("Runs dropped:", runsDropped)
}
