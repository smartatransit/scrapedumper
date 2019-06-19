package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/bipol/scrapedumper/pkg/dumper"
	"github.com/bipol/scrapedumper/pkg/martaapi"
	"go.uber.org/zap"
)

type WorkPoller interface {
	Poll(ctx context.Context, errC chan error) error
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ScrapeAndDumpClient
type ScrapeAndDumpClient struct {
	dumper          dumper.Dumper
	scheduleFinders []martaapi.ScheduleFinder
	pollTime        time.Duration
	logger          *zap.Logger
}

func New(dumper dumper.Dumper, pollTime time.Duration, logger *zap.Logger, apis ...martaapi.ScheduleFinder) ScrapeAndDumpClient {
	return ScrapeAndDumpClient{
		dumper,
		apis,
		pollTime,
		logger,
	}
}

func (c ScrapeAndDumpClient) Poll(ctx context.Context, errC chan error) {
	c.logger.Info("starting to poll")
	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("exiting poll")
				return
			default:
			}
			err := c.scrapeAndDump(ctx)
			if err != nil {
				errC <- err
			}
			time.Sleep(c.pollTime)
		}
	}()
}

func (c ScrapeAndDumpClient) scrapeAndDump(ctx context.Context) error {
	c.logger.Debug("scrape and dumping")
	for _, finder := range c.scheduleFinders {
		reader, err := finder.FindSchedules(ctx)
		if err != nil {
			return err
		}
		defer reader.Close()
		t := time.Now().UTC()
		path := fmt.Sprintf("%s.json", t.Format(time.RFC3339))
		err = c.dumper.Dump(ctx, reader, path)
		if err != nil {
			return err
		}
	}
	return nil
}
