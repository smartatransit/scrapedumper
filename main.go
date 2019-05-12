package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bipol/scrapedumper/pkg/dumper"
	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/bipol/scrapedumper/pkg/worker"
	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"
)

type options struct {
	S3BucketName      string `long:"s3-bucket-name" env:"S3_BUCKET_NAME" description:"s3 bucket to dump stuff into" required:"true"`
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

	awsSession := session.Must(session.NewSession())
	client := s3.New(awsSession)
	s3Manager := s3manager.NewUploaderWithClient(client)

	httpClient := http.Client{}

	martaClient := martaapi.New(&httpClient, opts.MartaAPIKey, logger)
	dump := dumper.New(s3Manager, opts.S3BucketName, logger)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	logger.Info(fmt.Sprintf("Poll time is %d seconds", opts.PollTimeInSeconds))
	poller := worker.New(dump, martaClient, time.Duration(opts.PollTimeInSeconds)*time.Second, logger)

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
