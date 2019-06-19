package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bipol/scrapedumper/pkg/dumper"
	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/bipol/scrapedumper/pkg/worker"
	"go.uber.org/zap"
)

type options struct {
	OutputLocation    string `long:"output-location" env:"OUTPUT_LOCATION" description:"local path to output" required:"true"`
	MartaAPIKey       string `long:"marta-api-key" env:"MARTA_API_KEY" description:"marta api key" required:"true"`
	PollTimeInSeconds int    `long:"poll-time-in-seconds" env:"POLL_TIME_IN_SECONDS" description:"time to poll marta api every second" required:"true"`
}

func main() {
	fmt.Println("Starting scrape and dump")
	var opts options
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any

	httpClient := http.Client{}

	trainClient := martaapi.New(&httpClient, opts.MartaAPIKey, logger, martaapi.RealtimeTrainTimeEndpoint, "train-data")
	busClient := martaapi.New(&httpClient, opts.MartaAPIKey, logger, martaapi.BusEndpoint, "bus-data")
	dump := dumper.New(opts.OutputLocation, logger)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	logger.Info(fmt.Sprintf("Poll time is %d seconds", opts.PollTimeInSeconds))
	poller := worker.New(dump, time.Duration(opts.PollTimeInSeconds)*time.Second, logger, trainClient, busClient)

	errC := make(chan error, 1)

	poller.Poll(ctx, errC)

	for err := range errC {
		logger.Error(err.Error())
	}
}
