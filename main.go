package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/smartatransit/scrapedumper/pkg/circuitbreaker"
	"github.com/smartatransit/scrapedumper/pkg/config"
	"github.com/smartatransit/scrapedumper/pkg/martaapi"
	"github.com/smartatransit/scrapedumper/pkg/worker"
)

type options struct {
	OutputLocation    string  `long:"output-location" env:"OUTPUT_LOCATION" description:"local path to output"`
	DynamoTableName   string  `long:"dynamo-table-name" env:"DYNAMO_TABLE_NAME" description:"dynamo table name"`
	S3BucketName      string  `long:"s3-bucket-name" env:"S3_BUCKET_NAME" description:"s3 bucket to dump stuff into"`
	MartaAPIKey       *string `long:"marta-api-key" env:"MARTA_API_KEY" description:"marta api key"`
	MartaAPIKeyFile   *string `long:"marta-api-key-file" env:"MARTA_API_KEY_FILE" description:"file containing the marta api key"`
	PollTimeInSeconds int     `long:"poll-time-in-seconds" env:"POLL_TIME_IN_SECONDS" description:"time to poll marta api every second" required:"true"`

	Debug      bool    `long:"debug" env:"DEBUG" description:"enabled debug logging"`
	ConfigPath *string `long:"config-path" env:"CONFIG_PATH" description:"An optional file that overrides the default configuration of sources and targets."`
}

func main() {
	fmt.Println("Starting scrape and dump")
	var opts options
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	var logger *zap.Logger
	if opts.Debug {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	defer func() {
		_ = logger.Sync() // flushes buffer, if any
	}()

	martaAPIKey := getMartaAPIKey(opts)

	wc, err := GetWorkConfig(opts)
	if err != nil {
		log.Fatal(err)
	}

	httpClient := http.Client{}

	trainClient := martaapi.New(&httpClient, martaAPIKey, logger, martaapi.RealtimeTrainTimeEndpoint, "train-data")
	busClient := martaapi.New(&httpClient, martaAPIKey, logger, martaapi.BusEndpoint, "bus-data")

	workList, cleanup, err := config.BuildWorkList(
		logger,
		sql.Open,
		wc,
		busClient,
		trainClient,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if cleanupErr := cleanup(); cleanupErr != nil {
			logger.Error(cleanupErr.Error())
		}
	}()

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

func getMartaAPIKey(opts options) string {
	if opts.MartaAPIKey != nil {
		return *opts.MartaAPIKey
	}

	if opts.MartaAPIKeyFile != nil {
		file, err := os.Open(*opts.MartaAPIKeyFile)
		if err != nil {
			log.Fatal(err)
		}

		bs, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}

		return string(bs)
	}

	log.Fatal("One of `--marta-api-key` or `--marta-api-key-file` is required")
	return ""
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
		err = errors.Wrapf(err, "failed opening config file %s for reading", *opts.ConfigPath)
		return
	}

	err = json.NewDecoder(file).Decode(&wc)
	err = errors.Wrapf(err, "failed parsing config file %s", file.Name())
	return
}

//BuildDefaultWorkConfig produces the default collection of dumpers
func BuildDefaultWorkConfig(opts options) config.WorkConfig {
	var (
		dumpConfig []config.DumpConfig
		busConfig  []config.DumpConfig
		cfg        config.WorkConfig
	)
	if opts.S3BucketName != "" {
		dumpConfig = append(dumpConfig,
			config.DumpConfig{
				Kind:         config.S3DumperKind,
				S3BucketName: opts.S3BucketName,
			},
		)
		busConfig = append(busConfig,
			config.DumpConfig{
				Kind:         config.S3DumperKind,
				S3BucketName: opts.S3BucketName,
			},
		)
	}
	if opts.OutputLocation != "" {
		dumpConfig = append(dumpConfig,
			config.DumpConfig{
				Kind:                config.FileDumperKind,
				LocalOutputLocation: opts.OutputLocation,
			},
		)
		busConfig = append(busConfig,
			config.DumpConfig{
				Kind:         config.S3DumperKind,
				S3BucketName: opts.S3BucketName,
			},
		)
	}
	if opts.DynamoTableName != "" {
		dumpConfig = append(dumpConfig,
			config.DumpConfig{
				Kind:            config.DynamoDBDumperKind,
				DynamoTableName: opts.DynamoTableName,
			},
		)
	}
	if len(dumpConfig) != 0 {
		cfg.TrainDumper = &config.DumpConfig{
			Kind:       config.RoundRobinKind,
			Components: dumpConfig,
		}
	}
	if len(busConfig) != 0 {
		cfg.BusDumper = &config.DumpConfig{
			Kind:       config.RoundRobinKind,
			Components: busConfig,
		}

	}
	return cfg
}
