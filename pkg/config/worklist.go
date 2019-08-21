package config

import (
	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/bipol/scrapedumper/pkg/worker"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//WorkConfig is the top-level config object, defined an entire
//scrapedumper job to be started.
type WorkConfig struct {
	BusDumper   *DumpConfig `json:"bus_dumper"`
	TrainDumper *DumpConfig `json:"train_dumper"`
}

//BuildWorkList builds a worklist from the specified clients
//and dumper config
func BuildWorkList(
	log *zap.Logger,
	sqlOpen SQLOpener,
	c WorkConfig,
	busClient martaapi.Client,
	trainClient martaapi.Client,
) (workList worker.WorkList, err error) {
	if c.BusDumper != nil {
		busDumper, err := BuildDumper(log, sqlOpen, *c.BusDumper)
		if err != nil {
			err = errors.Wrap(err, "failed to build bus dumper")
			return workList, err
		}
		workList.AddWork(busClient, busDumper)
	}

	if c.TrainDumper != nil {
		trainDumper, err := BuildDumper(log, sqlOpen, *c.TrainDumper)
		if err != nil {
			err = errors.Wrap(err, "failed to build train dumper")
			return workList, err
		}
		workList.AddWork(trainClient, trainDumper)
	}
	return
}
