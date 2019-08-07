package postgres

import (
	"time"

	"github.com/bipol/scrapedumper/pkg/marta"
	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/pkg/errors"
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
func (a *UpserterAgent) AddRecordToDatabase(rec martaapi.Schedule) (err error) {
	eventTime, err := time.Parse(martaapi.MartaAPIDatetimeFormat, rec.EventTime)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse record event time `%s`", rec.EventTime)
		return
	}

	runStartMoment, lastUpdated, err := a.repo.GetLatestRunStartMomentFor(marta.Direction(rec.Direction), marta.Line(rec.Line), rec.TrainID)
	if err != nil {
		err = errors.Wrapf(err, "failed to get latest run start moment for record `%s`", rec.String())
		return
	}

	//if the run didn't match, or if the latest run is stale,
	//then this is the start of a new run
	if runStartMoment == (time.Time{}) ||
		lastUpdated.Before(eventTime.Add(-a.runLifetime)) {

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
			err = errors.Wrapf(err, "failed to set arrival time from record `%s`", rec.String())
			return
		}
	} else {
		var estimate time.Time
		estimate, err = time.Parse(martaapi.MartaAPITimeFormat, rec.NextArrival)
		if err != nil {
			err = errors.Wrapf(err, "failed to parse record estimated arrival time `%s`", rec.NextArrival)
			return
		}

		//take the time part of estimate together with the date part of runStartMoment
		estimate = time.Date(
			runStartMoment.Year(), runStartMoment.Month(), runStartMoment.Day(),
			estimate.Hour(), estimate.Minute(), estimate.Second(), estimate.Nanosecond(),
			time.Local,
		)

		err = a.repo.AddArrivalEstimate(
			marta.Direction(rec.Direction),
			marta.Line(rec.Line),
			rec.TrainID,
			runStartMoment,
			marta.Station(rec.Station),
			ArrivalEstimate{
				EventTime:            eventTime,
				EstimatedArrivalTime: estimate,
			},
		)
		if err != nil {
			err = errors.Wrapf(err, "failed to add arrival estimate from record `%s`", rec.String())
			return
		}
	}

	return nil
}
