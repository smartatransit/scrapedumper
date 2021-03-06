package config

import (
	"github.com/smartatransit/scrapedumper/pkg/dumper"
	"github.com/smartatransit/scrapedumper/pkg/martaapi"
	"github.com/smartatransit/scrapedumper/pkg/worker"
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
) (workList worker.WorkList, f CleanupFunc, err error) {
	var cleanups []CleanupFunc
	var cleanup CleanupFunc
	if c.BusDumper != nil {
		var busDumper dumper.Dumper
		busDumper, cleanup, err = BuildDumper(log, sqlOpen, *c.BusDumper)
		if err != nil {
			err = errors.Wrap(err, "failed to build bus dumper")
			return
		}
		cleanups = append(cleanups, cleanup)
		workList.AddWork(busClient, busDumper)
	}

	if c.TrainDumper != nil {
		var trainDumper dumper.Dumper
		trainDumper, cleanup, err = BuildDumper(log, sqlOpen, *c.TrainDumper)
		if err != nil {
			err = errors.Wrap(err, "failed to build train dumper")
			return
		}
		cleanups = append(cleanups, cleanup)
		workList.AddWork(trainClient, trainDumper)
	}
	f = NewRoundRobinCleanup(cleanups)
	return
}
