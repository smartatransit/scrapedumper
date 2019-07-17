package postgres

import (
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartatransit/scrapedumper/pkg/marta"
)

//Repository implements interactions with Postgres through GORM
type Repository interface {
	EnsureTables() error

	GetLatestRunIdentifierFor(dir marta.Direction, line marta.Line, trainID string) (string, error)
	EnsureArrivalRecord(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station) error
	AddArrivalEstimate(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, estimate ArrivalEstimate) error
	SetArrivalTime(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, arrivalTime time.Time) error

	//TODO AddRecordToDatabase which accepts a time.Duration for de-duping
}

//RepositoryAgent implements Repository
type RepositoryAgent struct {
	DB *gorm.DB
}

//EnsureTables ensures that all necessary tables exist
func (a *RepositoryAgent) EnsureTables() error {
	return a.DB.AutoMigrate(&Arrival{}).Error
}

//GetLatestRunIdentifierFor gets the run identifier of the most recent run matching this info
func (a *RepositoryAgent) GetLatestRunIdentifierFor(dir marta.Direction, line marta.Line, trainID string) (string, error) {
	row := a.DB.Model(&Arrival{}).
		Where("direction = ?", string(dir)).
		Where("line = ?", string(line)).
		Where("train_id = ?", trainID).
		Order("run_first_event_moment DESC").Limit(1).
		Select("run_first_event_moment").Row()
	if a.DB.Error != nil {
		return "", a.DB.Error
	}

	if row == nil {
		return "", nil
	}

	var runFirstEventMoment time.Time
	err := row.Scan(&runFirstEventMoment)

	return fmt.Sprintf("%s_%s_%s_%s", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339)), err
}

//EnsureArrivalRecord ensures that a record exists for the specified arrival
func (a *RepositoryAgent) EnsureArrivalRecord(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station) error {
	err := a.DB.Set("gorm:insert_option", "ON CONFLICT DO NOTHING").
		Create(&Arrival{
			Direction:           dir,
			Line:                line,
			TrainID:             trainID,
			RunFirstEventMoment: runFirstEventMoment,
			Station:             station,
		}).Error
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil
		}
		return err
	}
	return nil
}

//AddArrivalEstimate upserts the specified arrival estimate to the arrival record in question
func (a *RepositoryAgent) AddArrivalEstimate(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, estimate ArrivalEstimate) error {
	if err := a.EnsureArrivalRecord(dir, line, trainID, runFirstEventMoment, station); err != nil {
		return nil
	}

	ident := IdentifierFor(dir, line, trainID, runFirstEventMoment, station)
	row := a.DB.Model(&Arrival{}).
		Where("identifier = ?", ident).
		Select("arrival_estimates").Row()
	if a.DB.Error != nil {
		return a.DB.Error
	}

	if row == nil {
		return errors.New("TODO")
	}

	var ests ArrivalEstimates
	err := row.Scan(&ests)
	if err != nil {
		return err
	}

	ests = append(ests, estimate)

	err = a.DB.Model(&Arrival{Identifier: ident}).
		Update("arrival_estimates", ests).Error
	if err != nil {
		return err
	}

	return nil
}

//SetArrivalTime upserts the specified actual arrival time to the arrival record in question
func (a *RepositoryAgent) SetArrivalTime(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, arrivalTime time.Time) error {
	if err := a.EnsureArrivalRecord(dir, line, trainID, runFirstEventMoment, station); err != nil {
		return nil
	}

	ident := IdentifierFor(dir, line, trainID, runFirstEventMoment, station)
	err := a.DB.Model(&Arrival{Identifier: ident}).
		Update("arrival_time", arrivalTime).Error
	if err != nil {
		return err
	}

	return nil
}

//TODO:
// - write a migrator and repository
// - write a BATCH LOADER interface
// - write a dumper interface
