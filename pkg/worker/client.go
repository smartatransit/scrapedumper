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

// ScrapeAndDumpClient contains all of the assets required to obtain scrape data from client, and write them to dump sites
type ScrapeAndDumpClient struct {
	workList WorkGetter
	pollTime time.Duration
	logger   *zap.Logger
}

func NewWorkList() *WorkList {
	return &WorkList{}
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . WorkGetter
type WorkGetter interface {
	GetWork() []ScrapeDump
}

func (w *WorkList) AddWork(sched martaapi.ScheduleFinder, dump dumper.Dumper) *WorkList {
	w.work = append(w.work, ScrapeDump{sched, dump})
	return w
}

func (w *WorkList) GetWork() []ScrapeDump {
	return w.work
}

// WorkList is a way to build up units of work (ScrapeDump work) in a way that allows us to pair a scrape with a dump
// this gives a user more freedom in what data gets dumped where
type WorkList struct {
	work []ScrapeDump
}

// ScrapeDump is a pairing of a client (scraper) and a dumper.
type ScrapeDump struct {
	Scraper martaapi.ScheduleFinder
	Dumper  dumper.Dumper
}

func New(pollTime time.Duration, logger *zap.Logger, workList WorkGetter) ScrapeAndDumpClient {
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
				c.logger.Error(err.Error())
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
