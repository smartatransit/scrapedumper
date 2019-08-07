package postgres

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/bipol/scrapedumper/pkg/marta"
)

//Repository implements interactions with Postgres through GORM
type Repository interface {
	EnsureTables() error

	GetLatestRunStartMomentFor(dir marta.Direction, line marta.Line, trainID string) (startTime time.Time, mostRecentEventTime time.Time, err error)
	EnsureArrivalRecord(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station) (err error)
	AddArrivalEstimate(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, eventTime time.Time, estimate time.Time) (err error)
	SetArrivalTime(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, eventTime time.Time, arrival time.Time) (err error)
}

//NewRepository creates a new postgres respotitory
func NewRepository(
	db *gorm.DB,
) *RepositoryAgent {
	return &RepositoryAgent{
		DB: db,
	}
}

//RepositoryAgent implements Repository
type RepositoryAgent struct {
	DB *gorm.DB
}

//EnsureTables ensures that all necessary tables exist
func (a *RepositoryAgent) EnsureTables() error {
	return a.DB.AutoMigrate(&Arrival{}).Error
}

//GetLatestRunStartMomentFor from among all runs matching the specified data, this function selects
//the most recent one and returns it's earliest start time (used as part of the run identifier) as
//well as it's most recent one, which is used for determining whether it is stale. If no runs match
//the metadata, it returns two zero time.Time objects and no error
func (a *RepositoryAgent) GetLatestRunStartMomentFor(dir marta.Direction, line marta.Line, trainID string) (startTime time.Time, mostRecentEventTime time.Time, err error) {
	var arr Arrival
	err = a.DB.Model(&Arrival{}).
		Where("direction = ?", string(dir)).
		Where("line = ?", string(line)).
		Where("train_id = ?", trainID).
		Order("run_first_event_moment DESC").
		Order("most_recent_event_time DESC").
		First(&arr).Error
	if err != nil {
		if err.Error() == "record not found" {
			err = nil
			return
		}
		err = errors.Wrapf(err, "failed to query latest run start moment for dir `%s` line `%s` and train `%s`", dir, line, trainID)
		return
	}

	startTime = arr.RunFirstEventMoment
	mostRecentEventTime = arr.MostRecentEventTime
	return
}

//EnsureArrivalRecord ensures that a record exists for the specified arrival
func (a *RepositoryAgent) EnsureArrivalRecord(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station) (err error) {
	err = a.DB.Set("gorm:insert_option", "ON CONFLICT DO NOTHING").
		Create(&Arrival{
			Direction:           dir,
			Line:                line,
			TrainID:             trainID,
			RunFirstEventMoment: runFirstEventMoment,
			Station:             station,
			ArrivalEstimates:    ArrivalEstimates(map[time.Time]time.Time{}),
		}).Error
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			err = nil
			return
		}
		err = errors.Wrapf(err, "failed to ensure arrival for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
		return
	}
	return nil
}

//AddArrivalEstimate upserts the specified arrival estimate to the arrival record in question
//NOTE: this method is NOT thread-safe.
func (a *RepositoryAgent) AddArrivalEstimate(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, eventTime time.Time, estimate time.Time) (err error) {
	if err = a.EnsureArrivalRecord(dir, line, trainID, runFirstEventMoment, station); err != nil {
		err = errors.Wrapf(err, "failed to ensure pre-existing arrival record for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
		return
	}

	var arr Arrival
	ident := IdentifierFor(dir, line, trainID, runFirstEventMoment, station)
	err = a.DB.Model(&Arrival{}).
		Where("identifier = ?", ident).
		First(&arr).Error
	if err != nil {
		if err.Error() == "record not found" {
			err = nil
			return
		}
		err = errors.Wrapf(err, "failed to get existing arrival estimates for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
		return
	}

	arr.ArrivalEstimates[eventTime] = estimate
	err = a.DB.Model(&Arrival{Identifier: ident}).
		Update("arrival_estimates", arr.ArrivalEstimates).
		Update("most_recent_event_time", eventTime).Error
	if err != nil {
		err = errors.Wrapf(err, "failed to add arrival estimate for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
		return
	}

	return
}

//SetArrivalTime upserts the specified actual arrival time to the arrival record in question
func (a *RepositoryAgent) SetArrivalTime(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, eventTime time.Time, arrivalTime time.Time) (err error) {
	if err = a.EnsureArrivalRecord(dir, line, trainID, runFirstEventMoment, station); err != nil {
		return
	}

	ident := IdentifierFor(dir, line, trainID, runFirstEventMoment, station)
	err = a.DB.Model(&Arrival{Identifier: ident}).
		// Where("arrival_time = ?", time.Time{}).
		Update("arrival_time", arrivalTime).
		Update("most_recent_event_time", eventTime).Error
	if err != nil {
		err = errors.Wrapf(err, "failed to set arrival time for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
		return err
	}

	return
}
