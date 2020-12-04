package worker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/smartatransit/scrapedumper/pkg/circuitbreaker"
	"github.com/smartatransit/scrapedumper/pkg/dumper"
	"github.com/smartatransit/scrapedumper/pkg/martaapi"
	"github.com/pkg/errors"
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
	cb       *circuitbreaker.CircuitBreaker
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

type Option = func(*ScrapeAndDumpClient)

func WithCircuitBreaker(c *circuitbreaker.CircuitBreaker) func(*ScrapeAndDumpClient) {
	return func(x *ScrapeAndDumpClient) {
		x.cb = c
	}
}

// New will initialize a new ScrapeDumper client, and if not provided with a circuit breaker, will fail immediately on the first error
//is is adviced to provide a circuitbreaker to manage this logic if you would rather this not occur
func New(pollTime time.Duration, logger *zap.Logger, workList WorkGetter, opts ...Option) ScrapeAndDumpClient {
	sc := ScrapeAndDumpClient{
		workList: workList,
		pollTime: pollTime,
		logger:   logger,
	}
	for _, opt := range opts {
		opt(&sc)
	}

	return sc
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
			var err error
			if c.cb != nil {
				err = c.cb.Run(func() error {
					innerErr := c.scrapeAndDumpAll(ctx)
					if innerErr != nil {
						c.logger.Error(innerErr.Error())
					}
					return innerErr
				})
				if err != nil && errors.Cause(err) == circuitbreaker.ErrSystemFailure {
					errC <- err
					return
				}
			} else {
				err := c.scrapeAndDumpAll(ctx)
				if err != nil {
					errC <- err
					return
				}
			}
			time.Sleep(c.pollTime)
		}
	}()
}

func (c ScrapeAndDumpClient) scrapeAndDumpAll(ctx context.Context) (err error) {
	c.logger.Debug("scrape and dumping")
	for _, sd := range c.workList.GetWork() {
		if sdErr := c.scrapeAndDump(ctx, sd); sdErr != nil {
			err = sdErr
		}
	}
	return err
}

func (c ScrapeAndDumpClient) scrapeAndDump(ctx context.Context, sd ScrapeDump) (err error) {
	var reader io.ReadCloser
	reader, err = sd.Scraper.FindSchedules(ctx)
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
	return nil
}
