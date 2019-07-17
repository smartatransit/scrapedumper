package config

import (
	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/bipol/scrapedumper/pkg/worker"
	"go.uber.org/zap"
)

//WorkConfig is the top-level config object, defined an entire
//scrapedumper job to be started.
type WorkConfig struct {
	BusDumper   DumpConfig `json:"bus_dumper"`
	TrainDumper DumpConfig `json:"train_dumper"`
}

//BuildWorkList builds a worklist from the specified clients
//and dumper config
func BuildWorkList(
	log *zap.Logger,
	c WorkConfig,
	busClient martaapi.Client,
	trainClient martaapi.Client,
) (workList worker.WorkList, err error) {
	busDumper, err := BuildDumper(log, c.BusDumper)
	if err != nil {
		return
	}

	trainDumper, err := BuildDumper(log, c.TrainDumper)
	if err != nil {
		return
	}

	workList.
		AddWork(trainClient, trainDumper).
		AddWork(busClient, busDumper)

	return
}
