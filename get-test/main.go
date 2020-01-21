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

//
//
// TODO remove this entire folder
//
//

type options struct {
	PostgresConnectionString string `long:"postgres-connection-string" env:"POSTGRES_CONNECTION_STRING" required:"true"`
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

	threshold := time.Now().Add(-5 * time.Minute)
	repo := postgres.NewRepository(logger, db)
	for {
		nextThresh := time.Now()
		runs, err := repo.GetRecentlyActiveRuns(postgres.EasternTime(threshold))
		if err != nil {
			logger.Error(err.Error())
		}

		logger.Info(fmt.Sprintf("%v changes", len(runs)))

		threshold = nextThresh
		time.Sleep(15 * time.Second)
	}
}
