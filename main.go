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
	"github.com/bipol/scrapedumper/pkg/circuitbreaker"
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
	busClient := martaapi.New(&httpClient, opts.MartaAPIKey, logger, martaapi.BusEndpoint, "bus-data")

	var trainDumps []dumper.Dumper
	var busDumps []dumper.Dumper
	if opts.S3BucketName != "" {
		logger.Info(fmt.Sprintf("activating s3 dumper %s", opts.S3BucketName))
		s3Dump := dumper.NewS3DumpHandler(s3Manager, opts.S3BucketName, logger)
		trainDumps = append(trainDumps, s3Dump)
		busDumps = append(busDumps, s3Dump)
	}
	if opts.OutputLocation != "" {
		logger.Info(fmt.Sprintf("activating local dumper %s", opts.OutputLocation))
		localDump := dumper.NewLocalDumpHandler(opts.OutputLocation, logger, afero.NewOsFs())
		trainDumps = append(trainDumps, localDump)
		busDumps = append(busDumps, localDump)
	}
	if opts.DynamoTableName != "" {
		logger.Info(fmt.Sprintf("activating dynamo dumper %s", opts.DynamoTableName))
		dynamoDump := dumper.NewDynamoDumpHandler(logger, opts.DynamoTableName, svc, martaapi.DigestScheduleResponse)
		trainDumps = append(trainDumps, dynamoDump)
	}

	trainRoundRobinDumper := dumper.NewRoundRobinDumpClient(logger, trainDumps...)
	busRoundRobinDumper := dumper.NewRoundRobinDumpClient(logger, busDumps...)

	var workList worker.WorkList
	workList.AddWork(trainClient, trainRoundRobinDumper).AddWork(busClient, busRoundRobinDumper)

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
