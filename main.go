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
)

type options struct {
	S3BucketName string `long:"s3-bucket-name" env:"S3_BUCKET_NAME" description:"s3 bucket to dump stuff into" required:"true"`
	MartaAPIKey  string `long:"marta-api-key" env:"MARTA_API_KEY" description:"marta api key" required:"true"`
}

func main() {
	fmt.Println("Starting scrape and dump")
	var opts options
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	awsSession := session.Must(session.NewSession())
	client := s3.New(awsSession)
	s3Manager := s3manager.NewUploaderWithClient(client)

	httpClient := http.Client{}

	martaClient := martaapi.New(&httpClient, opts.MartaAPIKey)
	dump := dumper.New(s3Manager, opts.S3BucketName)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	errC := make(chan error, 1)

	poller := worker.New(dump, martaClient, 15*time.Second)
	poller.Poll(ctx, errC)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	select {
	case err := <-errC:
		fmt.Println(err.Error())
		fmt.Println("shutting down...")
	case <-quit:
		cancelFunc()
		fmt.Println("interrupt signal received")
		fmt.Println("shutting down...")
	}

}
