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
	OutputLocation    string `long:"output-location" env:"OUTPUT_LOCATION" description:"local path to output"`
	DynamoTableName   string `long:"dynamo-table-name" env:"DYNAMO_TABLE_NAME" description:"dynamo table name"`
	S3BucketName      string `long:"s3-bucket-name" env:"S3_BUCKET_NAME" description:"s3 bucket to dump stuff into"`
	MartaAPIKey       string `long:"marta-api-key" env:"MARTA_API_KEY" description:"marta api key" required:"true"`
	PollTimeInSeconds int    `long:"poll-time-in-seconds" env:"POLL_TIME_IN_SECONDS" description:"time to poll marta api every second" required:"true"`

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

	wc, err := GetWorkConfig(opts)
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
func GetWorkConfig(opts options) (wc config.WorkConfig, err error) {
	if opts.ConfigPath == nil {
		wc = BuildDefaultWorkConfig(opts)
		return
	}

	file, err := os.Open(*opts.ConfigPath)
	if err != nil {
		return
	}

	err = json.NewDecoder(file).Decode(&wc)
	return
}

//BuildDefaultWorkConfig produces the default collection of dumpers
func BuildDefaultWorkConfig(opts options) config.WorkConfig {
	return config.WorkConfig{
		TrainDumper: config.DumpConfig{
			Kind: config.RoundRobinKind,
			Components: []config.DumpConfig{
				config.DumpConfig{
					Kind:         config.S3DumperKind,
					S3BucketName: opts.S3BucketName,
				},
				config.DumpConfig{
					Kind:                config.FileDumperKind,
					LocalOutputLocation: opts.OutputLocation,
				},
				config.DumpConfig{
					Kind:            config.DynamoDBDumperKind,
					DynamoTableName: opts.DynamoTableName,
				},
			},
		},
		BusDumper: config.DumpConfig{
			Kind: config.RoundRobinKind,
			Components: []config.DumpConfig{
				config.DumpConfig{
					Kind:         config.S3DumperKind,
					S3BucketName: opts.S3BucketName,
				},
				config.DumpConfig{
					Kind:                config.FileDumperKind,
					LocalOutputLocation: opts.OutputLocation,
				},
			},
		},
	}
}
