package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"

	"github.com/bipol/scrapedumper/pkg/circuitbreaker"
	"github.com/bipol/scrapedumper/pkg/config"
	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/bipol/scrapedumper/pkg/worker"
)

type options struct {
	config.GlobalConfig

	ConfigPath *string `long:"config-path" env:"CONFIG_PATH" description:"An optional file that overrides the default configuration of sources and targets."`
}

func main() {
	fmt.Println("Starting scrape and dump")
	var opts options
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync() // flushes buffer, if any
	}()

	wc, err := GetWorkConfig(opts.ConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	httpClient := http.Client{}

	trainClient := martaapi.New(&httpClient, opts.MartaAPIKey, logger, martaapi.RealtimeTrainTimeEndpoint, "train-data")
	busClient := martaapi.New(&httpClient, opts.MartaAPIKey, logger, martaapi.BusEndpoint, "bus-data")

	workList, err := config.BuildWorkList(
		logger,
		wc,
		busClient,
		trainClient,
		opts.GlobalConfig,
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	cb := circuitbreaker.New(logger, 1*time.Hour, 10)

	logger.Info(fmt.Sprintf("Poll time is %d seconds", opts.PollTimeInSeconds))
	poller := worker.New(time.Duration(opts.PollTimeInSeconds)*time.Second, logger, &workList, worker.WithCircuitBreaker(cb))

	errC := make(chan error, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	poller.Poll(ctx, errC)

	select {
	case err := <-errC:
		logger.Error(err.Error())
		logger.Info("shutting down...")
	case <-quit:
		cancelFunc()
		logger.Info("interrupt signal received")
		logger.Info("shutting down...")
	}

}

//GetWorkConfig gets the WorkConfig either from a JSON file or from
//the hard-coded default.
func GetWorkConfig(path *string) (wc config.WorkConfig, err error) {
	if path == nil {
		wc = config.DefaultWorkConfig
		return
	}

	file, err := os.Open(*path)
	if err != nil {
		return
	}

	err = json.NewDecoder(file).Decode(&wc)
	return
}
