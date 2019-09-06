package config

import (
	"github.com/bipol/scrapedumper/pkg/dumper"
	"github.com/bipol/scrapedumper/pkg/worker"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ScraperKind string

const (
	WebKind = "WEB"
)

//ErrScraperValidationFailed indicates that a scraper's configuration was invalid
var ErrScraperValidationFailed = errors.New("scraper failed to build due to missing args")

type ScrapeConfig struct {
	Kind    ScraperKind       `json:"kind"`
	Options map[string]string `json:"options"`
	Dumper  DumpConfig        `json:"dumper"`
}

func BuildScraper(log *zap.Logger, s ScrapeConfig) (worker.Scraper, dumper.Dumper, error) {
	switch s.Kind {
	case WebKind:
		dumper, err := BuildDumper(log, s.Dumper)
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, errors.Wrapf(ErrScraperValidationFailed, "unsupported scraper kind `%s`", string(s.Kind))
	}
}
