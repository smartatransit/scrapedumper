package postgres

import (
	"time"

	"github.com/bipol/scrapedumper/pkg/marta"
	"github.com/bipol/scrapedumper/pkg/martaapi"
)

//Upserter upserts a record to the database, while attempting to
//reconcile separate records from the same train run
type Upserter interface {
	AddRecordToDatabase(martaapi.Schedule) error
}

//UpserterAgent implements Upserter
type UpserterAgent struct {
	Repo        Repository
	RunLifetime time.Duration
}

//AddRecordToDatabase upserts a record to the database, while
//attempting to reconcile separate records from the same train run
func (a *UpserterAgent) AddRecordToDatabase(rec martaapi.Schedule) error {
	runStartMoment, lastUpdated, err := a.Repo.GetLatestRunStartMomentFor(marta.Direction(rec.Direction), marta.Line(rec.Line), rec.TrainID)
	if err != nil {
		return err
	}

	eventTime, err := time.Parse(martaapi.MartaAPITimeFormat, rec.EventTime)
	if err != nil {
		return err
	}

	//if the run didn't match, or if the latest run is stale,
	//then this is the start of a new run
	if runStartMoment == (time.Time{}) ||
		lastUpdated.Before(time.Now().Add(-a.RunLifetime)) {

		runStartMoment = eventTime
	}

	if rec.HasArrived() {
		err = a.Repo.SetArrivalTime(marta.Direction(rec.Direction), marta.Line(rec.Line), rec.TrainID, runStartMoment, marta.Station(rec.Station), eventTime)
		if err != nil {
			return err
		}
	} else {
		estimate, err := time.Parse(martaapi.MartaAPITimeFormat, rec.NextArrival)
		if err != nil {
			return err
		}

		err = a.Repo.AddArrivalEstimate(
			marta.Direction(rec.Direction), marta.Line(rec.Line), rec.TrainID, runStartMoment, marta.Station(rec.Station),
			ArrivalEstimate{
				EventTime:            eventTime,
				EstimatedArrivalTime: estimate,
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}
