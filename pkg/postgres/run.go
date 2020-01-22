package postgres

import (
	"strings"

	"github.com/bipol/scrapedumper/pkg/martaapi"
)

type Run struct {
	Identifier            string `json:"identifier"`
	RunGroupIdentifier    string `json:"run_group_identifier"`
	CorrectedLine         string `json:"corrected_line"`
	CorrectedDirection    string `json:"corrected_direction"`
	MostRecentEventMoment string `json:"most_recent_event_moment"`
	RunFirstEventMoment   string `json:"run_first_event_moment"`

	Line      martaapi.Line      `json:"line"`
	Direction martaapi.Direction `json:"direction"`
	TrainID   string             `json:"train_id"`

	Arrivals Arrivals `json:"arrivals"`
}

func (r Run) setLineDirectionAndTrainID() {
	parts := strings.Split(r.Identifier, "_")

	r.Line = martaapi.Line(parts[0])
	r.Direction = martaapi.Direction(parts[1])
	r.TrainID = parts[2]
}

func (r Run) Finished() bool {
	setLineDirectionAndTrainID()
	terminus := martaapi.Termini[run.Line][run.Direction]
	return run.Arrivals[terminus].ArrivalTime != nil
}

type Arrivals map[martaapi.Station]Arrival

type Arrival struct {
	Identifier  string           `json:"identifier"`
	Station     martaapi.Station `json:"station"`
	ArrivalTime *EasternTime     `json:"arrival_time"`

	Estimates EstimateList `json:"estimates"`
}

type EstimateList map[EasternTime]EasternTime
