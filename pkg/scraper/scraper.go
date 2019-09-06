package scraper

import (
	"context"
	"io"
	"net/http"

	"github.com/bipol/scrapedumper/pkg/martaapi"
	"go.uber.org/zap"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Doer
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ScheduleFinder
type ScheduleFinder interface {
	FindSchedules(ctx context.Context) (io.ReadCloser, error)
	Prefix() string
}

type MartaTrainScrapeClient struct {
	logger      *zap.Logger
	trainClient ScheduleFinder
}

func NewMartaTrainScrapeClient(logger *zap.Logger, apiKey string, doer Doer) MartaTrainScrapeClient {
	trainClient := martaapi.New(doer, apiKey, logger, martaapi.RealtimeTrainTimeEndpoint, "train-data")

	return MartaTrainScrapeClient{
		logger,
		trainClient,
	}
}

func (c MartaTrainScrapeClient) Scrape(ctx context.Context) (io.ReadCloser, error) {
	return c.trainClient.FindSchedules(ctx)
}

type MartaBusScrapeClient struct {
	logger    *zap.Logger
	busClient ScheduleFinder
}

func NewMartaBusScrapeClient(logger *zap.Logger, apiKey string, doer Doer) MartaBusScrapeClient {
	busClient := martaapi.New(doer, apiKey, logger, martaapi.BusEndpoint, "bus-data")
	return MartaBusScrapeClient{
		logger,
		busClient,
	}
}

func (c MartaBusScrapeClient) Scrape(ctx context.Context) (io.ReadCloser, error) {
	return c.busClient.FindSchedules(ctx)
}
