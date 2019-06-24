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
	workList WorkList
	pollTime time.Duration
	logger   *zap.Logger
}

func NewWorkList() *WorkList {
	return &WorkList{}
}

func (w *WorkList) AddWork(sched martaapi.ScheduleFinder, dump dumper.Dumper) *WorkList {
	w.work = append(w.work, ScrapeDump{sched, dump})
	return w
}

func (w *WorkList) GetWork() []ScrapeDump {
	return w.work
}

type WorkList struct {
	work []ScrapeDump
}

type ScrapeDump struct {
	Scraper martaapi.ScheduleFinder
	Dumper  dumper.Dumper
}

func New(pollTime time.Duration, logger *zap.Logger, workList WorkList) ScrapeAndDumpClient {
	return ScrapeAndDumpClient{
		workList,
		pollTime,
		logger,
	}
}

func (c ScrapeAndDumpClient) Poll(ctx context.Context, errC chan error) {
	c.logger.Info("starting to poll")
	go func() {
		defer close(errC)
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
				return
			}
			time.Sleep(c.pollTime)
		}
	}()
}

func (c ScrapeAndDumpClient) scrapeAndDump(ctx context.Context) error {
	c.logger.Debug("scrape and dumping")
	for _, sd := range c.workList.GetWork() {
		reader, err := sd.Scraper.FindSchedules(ctx)
		if err != nil {
			return err
		}
		defer reader.Close()
		t := time.Now().UTC()
		path := fmt.Sprintf("%s/%s.json", sd.Scraper.Prefix(), t.Format(time.RFC3339))
		err = sd.Dumper.Dump(ctx, reader, path)
		if err != nil {
			return err
		}
	}
	return nil
}
