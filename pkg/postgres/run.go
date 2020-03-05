package postgres

import (
	"github.com/smartatransit/scrapedumper/pkg/martaapi"
)

type Run struct {
	Identifier            string             `json:"identifier"`
	RunGroupIdentifier    string             `json:"run_group_identifier"`
	CorrectedLine         martaapi.Line      `json:"line"`
	CorrectedDirection    martaapi.Direction `json:"direction"`
	MostRecentEventMoment string             `json:"most_recent_event_moment"`
	RunFirstEventMoment   string             `json:"run_first_event_moment"`

	Arrivals Arrivals `json:"arrivals"`
}

func (r Run) Finished() bool {
	terminus := martaapi.Termini[r.CorrectedLine][r.CorrectedDirection]
	return r.Arrivals[terminus].ArrivalTime != nil
}

type Arrivals map[martaapi.Station]Arrival

type Arrival struct {
	Identifier  string           `json:"identifier"`
	Station     martaapi.Station `json:"station"`
	ArrivalTime *EasternTime     `json:"arrival_time"`

	Estimates EstimateList `json:"estimates"`
}

type EstimateList map[EasternTime]EasternTime
