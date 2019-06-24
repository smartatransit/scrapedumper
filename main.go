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
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bipol/scrapedumper/pkg/dumper"
	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/bipol/scrapedumper/pkg/worker"
	"github.com/jessevdk/go-flags"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

type options struct {
	OutputLocation    string `long:"output-location" env:"OUTPUT_LOCATION" description:"local path to output"`
	DynamoTableName   string `long:"dynamo-table-name" env:"DYNAMO_TABLE_NAME" description:"dynamo table name"`
	S3BucketName      string `long:"s3-bucket-name" env:"S3_BUCKET_NAME" description:"s3 bucket to dump stuff into"`
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
	svc := dynamodb.New(awsSession)

	httpClient := http.Client{}

	trainClient := martaapi.New(&httpClient, opts.MartaAPIKey, logger, martaapi.RealtimeTrainTimeEndpoint, "train-data")

	//unfortunately, the dynamo dymp handler's marshal func does not currently account for bus client stuff.. this needs to be fixed
	//busClient := martaapi.New(&httpClient, opts.MartaAPIKey, logger, martaapi.BusEndpoint, "bus-data")

	var dumpClients []dumper.Dumper
	if opts.S3BucketName != "" {
		logger.Info(fmt.Sprintf("activating s3 dumper %s", opts.S3BucketName))
		s3Dump := dumper.NewS3DumpHandler(s3Manager, opts.S3BucketName, logger)
		dumpClients = append(dumpClients, s3Dump)
	}
	if opts.OutputLocation != "" {
		logger.Info(fmt.Sprintf("activating local dumper %s", opts.OutputLocation))
		localDump := dumper.NewLocalDumpHandler(opts.OutputLocation, logger, afero.NewOsFs())
		dumpClients = append(dumpClients, localDump)
	}
	if opts.DynamoTableName != "" {
		logger.Info(fmt.Sprintf("activating dynamo dumper %s", opts.DynamoTableName))
		dynamoDump := dumper.NewDynamoDumpHandler(logger, opts.DynamoTableName, svc, martaapi.DigestScheduleResponse)
		dumpClients = append(dumpClients, dynamoDump)
	}

	if len(dumpClients) == 0 {
		logger.Error("must specify a dump client")
		return
	}

	dump := dumper.NewRoundRobinDumpClient(logger, dumpClients...)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	logger.Info(fmt.Sprintf("Poll time is %d seconds", opts.PollTimeInSeconds))
	poller := worker.New(dump, time.Duration(opts.PollTimeInSeconds)*time.Second, logger, trainClient)

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
