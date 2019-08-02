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

//NewUpserter creates a new postgres upserter
func NewUpserter(
	repo Repository,
	runLifetime time.Duration,
) *UpserterAgent {
	return &UpserterAgent{
		repo:        repo,
		runLifetime: runLifetime,
	}
}

//UpserterAgent implements Upserter
type UpserterAgent struct {
	repo        Repository
	runLifetime time.Duration
}

//AddRecordToDatabase upserts a record to the database, while
//attempting to reconcile separate records from the same train run
func (a *UpserterAgent) AddRecordToDatabase(rec martaapi.Schedule) error {
	runStartMoment, lastUpdated, err := a.repo.GetLatestRunStartMomentFor(marta.Direction(rec.Direction), marta.Line(rec.Line), rec.TrainID)
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
		lastUpdated.Before(time.Now().Add(-a.runLifetime)) {

		runStartMoment = eventTime
	}

	if rec.HasArrived() {
		err = a.repo.SetArrivalTime(
			marta.Direction(rec.Direction),
			marta.Line(rec.Line),
			rec.TrainID,
			runStartMoment,
			marta.Station(rec.Station),
			ArrivalEstimate{
				EventTime:            eventTime,
				EstimatedArrivalTime: time.Time{}, /* TODO */
			},
		)
		if err != nil {
			return err
		}
	} else {
		estimate, err := time.Parse(martaapi.MartaAPITimeFormat, rec.NextArrival)
		if err != nil {
			return err
		}

		err = a.repo.AddArrivalEstimate(
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
