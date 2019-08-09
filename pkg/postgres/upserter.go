package postgres

import (
	"time"

	"github.com/bipol/scrapedumper/pkg/marta"
	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/pkg/errors"
)

//Upserter upserts a record to the database, while attempting to
//reconcile separate records from the same train run
//go:generate counterfeiter . Upserter
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
	eventTime, err := time.ParseInLocation(martaapi.MartaAPIDatetimeFormat, rec.EventTime, EasternTime)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse record event time `%s`", rec.EventTime)
		return
	}

	runStartMoment, mostRecentEventTime, err := a.repo.GetLatestRunStartMomentFor(marta.Direction(rec.Direction), marta.Line(rec.Line), rec.TrainID)
	if err != nil {
		err = errors.Wrapf(err, "failed to get latest run start moment for record `%s`", rec.String())
		return
	}

	//if the run didn't match, or if the latest run is stale,
	//then this is the start of a new run
	if runStartMoment == (time.Time{}) ||
		mostRecentEventTime.Before(eventTime.Add(-a.runLifetime)) {

		runStartMoment = eventTime
	}

	if err = a.repo.EnsureArrivalRecord(
		marta.Direction(rec.Direction),
		marta.Line(rec.Line),
		rec.TrainID,
		runStartMoment,
		marta.Station(rec.Station),
	); err != nil {
		err = errors.Wrapf(err, "failed to ensure pre-existing arrival record for `%s`", rec.String())
		return
	}

	if rec.HasArrived() {
		//TODO this is a potential source of error - there may be smarter ways to infer the arrival moment
		arrivalTime := eventTime
		err = a.repo.SetArrivalTime(
			marta.Direction(rec.Direction),
			marta.Line(rec.Line),
			rec.TrainID,
			runStartMoment,
			marta.Station(rec.Station),
			eventTime,
			arrivalTime,
		)
		if err != nil {
			err = errors.Wrapf(err, "failed to set arrival time from record `%s`", rec.String())
			return
		}
	} else {
		var estimate time.Time
		estimate, err = time.ParseInLocation(martaapi.MartaAPITimeFormat, rec.NextArrival, EasternTime)
		if err != nil {
			err = errors.Wrapf(err, "failed to parse record estimated arrival time `%s`", rec.NextArrival)
			return
		}

		//take the time part of estimate together with the date part of runStartMoment
		estimate = time.Date(
			runStartMoment.Year(), runStartMoment.Month(), runStartMoment.Day(),
			estimate.Hour(), estimate.Minute(), estimate.Second(), estimate.Nanosecond(),
			EasternTime,
		)

		err = a.repo.AddArrivalEstimate(
			marta.Direction(rec.Direction),
			marta.Line(rec.Line),
			rec.TrainID,
			runStartMoment,
			marta.Station(rec.Station),
			eventTime,
			estimate,
		)
		if err != nil {
			err = errors.Wrapf(err, "failed to add arrival estimate from record `%s`", rec.String())
			return
		}
	}

	return nil
}
