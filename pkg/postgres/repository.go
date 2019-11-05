package postgres

import (
	"database/sql"
	"fmt"

	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//Repository implements interactions with Postgres through GORM
//go:generate counterfeiter . Repository
type Repository interface {
	EnsureTables() error

	GetLatestRunStartMomentFor(dir martaapi.Direction, line martaapi.Line, trainID string, asOfMoment EasternTime) (runFirstEventMoment EasternTime, mostRecentEventTime EasternTime, err error)
	CreateRunRecord(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime) (err error)
	EnsureArrivalRecord(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, station martaapi.Station) (err error)
	AddArrivalEstimate(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, station martaapi.Station, eventTime EasternTime, estimate EasternTime) (err error)
	SetArrivalTime(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, station martaapi.Station, eventTime EasternTime, arrival EasternTime) (err error)
}

//NewRepository creates a new postgres respotitory
func NewRepository(
	logger *zap.Logger,
	db *sql.DB,
) *RepositoryAgent {
	return &RepositoryAgent{
		Logger: logger,
		DB:     db,
	}
}

//RepositoryAgent implements Repository
type RepositoryAgent struct {
	Logger *zap.Logger
	DB     *sql.DB
}

//EstimateIdentifierFor creates a identifier for the given metadata
func EstimateIdentifierFor(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, station martaapi.Station, estimateMoment EasternTime) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s_%s", dir, line, trainID, runFirstEventMoment.String(), station, estimateMoment)
}

//ArrivalIdentifierFor creates a identifier for the given metadata
func ArrivalIdentifierFor(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, station martaapi.Station) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", dir, line, trainID, runFirstEventMoment.String(), station)
}

//RunIdentifierFor creates a run identifier for the given metadata
func RunIdentifierFor(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime) string {
	return fmt.Sprintf("%s_%s_%s_%s", dir, line, trainID, runFirstEventMoment.String())
}

//RunGroupIdentifierFor creates a run group identifier for the given metadata
func RunGroupIdentifierFor(dir martaapi.Direction, line martaapi.Line, trainID string) string {
	return fmt.Sprintf("%s_%s_%s", dir, line, trainID)
}

//EnsureTables ensures that all necessary tables exist
func (a *RepositoryAgent) EnsureTables() error {
	_, err := a.DB.Exec(`
CREATE TABLE IF NOT EXISTS runs
(	identifier text,
	run_group_identifier text NOT NULL,
	most_recent_event_moment varchar NOT NULL,
	run_first_event_moment varchar NOT NULL,
	PRIMARY KEY (identifier)
)`)
	if err != nil {
		return errors.Wrapf(err, "failed to ensure runs table")
	}

	_, err = a.DB.Exec(`
CREATE TABLE IF NOT EXISTS arrivals
(	identifier text,
	run_identifier text NOT NULL,
	station text NOT NULL,
	arrival_time varchar,
	PRIMARY KEY (identifier)
)`)
	if err != nil {
		return errors.Wrapf(err, "failed to ensure arrivals table")
	}

	_, err = a.DB.Exec(`
CREATE TABLE IF NOT EXISTS estimates
(	identifier text,
	run_identifier text NOT NULL,
	arrival_identifier text NOT NULL,
	estimate_moment varchar NOT NULL,
	estimated_arrival_time varchar NOT NULL,
	PRIMARY KEY (identifier)
)`)
	if err != nil {
		return errors.Wrapf(err, "failed to ensure estimates table")
	}

	_, err = a.DB.Exec(`CREATE INDEX ON runs USING btree(run_group_identifier)`)
	if err != nil {
		return errors.Wrapf(err, "failed to index runs by run group")
	}
	_, err = a.DB.Exec(`CREATE INDEX ON arrivals USING btree(run_identifier)`)
	if err != nil {
		return errors.Wrapf(err, "failed to index arrivals by run")
	}
	_, err = a.DB.Exec(`CREATE INDEX ON estimates USING btree(arrival_identifier)`)
	if err != nil {
		return errors.Wrapf(err, "failed to index estimates by arrival")
	}
	_, err = a.DB.Exec(`CREATE INDEX ON estimates USING btree(run_identifier)`)
	if err != nil {
		return errors.Wrapf(err, "failed to index estimates by run")
	}

	_, err = a.DB.Exec(`
CREATE INDEX ON runs USING btree(
	run_group_identifier,
	run_first_event_moment DESC,
	most_recent_event_moment DESC
)`)
	return errors.Wrap(err, "failed to index runs for upserting")
}

//GetLatestRunStartMomentFor from among all runs in this run group, this method selects the most recently
//updated one and one and returns it's earliest event moment (used as part of the run identifier) as
//well as it's most recent one. If no runs are in the run_group, it returns two zero time.Time objects
//and no error.
func (a *RepositoryAgent) GetLatestRunStartMomentFor(dir martaapi.Direction, line martaapi.Line, trainID string, asOfMoment EasternTime) (runFirstEventMoment EasternTime, mostRecentEventTime EasternTime, err error) {
	row := a.DB.QueryRow(`
SELECT run_first_event_moment, runs.most_recent_event_moment
FROM arrivals JOIN runs ON runs.identifier = arrivals.run_identifier
WHERE run_group_identifier = $1 AND runs.most_recent_event_moment <= $2
ORDER BY run_first_event_moment DESC, runs.most_recent_event_moment DESC, arrivals.identifier ASC
LIMIT 1`,
		RunGroupIdentifierFor(dir, line, trainID),
		asOfMoment,
	)

	err = row.Scan(&runFirstEventMoment, &mostRecentEventTime)
	if err == sql.ErrNoRows {
		err = nil
		return
	}
	err = errors.Wrapf(err, "failed to query latest run start moment for dir `%s` line `%s` and train `%s`", dir, line, trainID)
	return
}

//CreateRunRecord inserts this run to the run table
func (a *RepositoryAgent) CreateRunRecord(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime) (err error) {
	res, err := a.DB.Exec(`
INSERT INTO runs
(identifier, run_group_identifier, most_recent_event_moment, run_first_event_moment)
VALUES ($1, $2, $3, $4)`,
		RunIdentifierFor(dir, line, trainID, runFirstEventMoment),
		RunGroupIdentifierFor(dir, line, trainID),
		runFirstEventMoment, //most_recent_event_moment
		runFirstEventMoment,
	)
	if err != nil {
		err = errors.Wrapf(err, "failed to create run for dir `%s` line `%s` train `%s` and first event moment `%s`", dir, line, trainID, runFirstEventMoment)
		return
	}

	i, err := res.RowsAffected()
	if err != nil {
		err = errors.Wrapf(err, "received malformed result when creating run for dir `%s` line `%s` train `%s` and first event moment `%s`", dir, line, trainID, runFirstEventMoment)
		return
	}
	if i != 1 {
		err = fmt.Errorf("create-run query unexpectedly affected %v rows - expected 1", i)
		return
	}

	return
}

//EnsureArrivalRecord ensures that a record exists for the specified arrival
func (a *RepositoryAgent) EnsureArrivalRecord(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, station martaapi.Station) (err error) {
	_, err = a.DB.Exec(`
INSERT INTO arrivals
(identifier, run_identifier, station)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING`,
		ArrivalIdentifierFor(dir, line, trainID, runFirstEventMoment, station),
		RunIdentifierFor(dir, line, trainID, runFirstEventMoment),
		station,
	)
	err = errors.Wrapf(err, "failed to ensure arrival for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment, station)
	return
}

//AddArrivalEstimate creates the specified arrival estimate
func (a *RepositoryAgent) AddArrivalEstimate(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, station martaapi.Station, eventTime EasternTime, estimate EasternTime) (err error) {
	tx, err := a.DB.Begin()
	if err != nil {
		err = errors.Wrapf(err, "failed to begin transaction to add arrival estimate for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment, station)
		return
	}

	_, err = tx.Exec(`
INSERT INTO estimates
(identifier, run_identifier, arrival_identifier, estimate_moment, estimated_arrival_time)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT DO NOTHING`,
		EstimateIdentifierFor(dir, line, trainID, runFirstEventMoment, station, eventTime),
		RunIdentifierFor(dir, line, trainID, runFirstEventMoment),
		ArrivalIdentifierFor(dir, line, trainID, runFirstEventMoment, station),
		eventTime,
		estimate,
	)
	if err != nil {
		rollback(tx, a.Logger)
		err = errors.Wrapf(err, "failed to add arrival estimate for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment, station)
		return
	}

	err = touchRun(tx, dir, line, trainID, runFirstEventMoment, eventTime)
	if err != nil {
		rollback(tx, a.Logger)
		return
	}

	err = errors.Wrapf(tx.Commit(), "failed to commit transaction when setting arrival time for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.String(), station)
	return
}

//SetArrivalTime upserts the specified actual arrival time to the arrival record in question
func (a *RepositoryAgent) SetArrivalTime(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, station martaapi.Station, eventTime EasternTime, arrivalTime EasternTime) (err error) {
	tx, err := a.DB.Begin()
	if err != nil {
		err = errors.Wrapf(err, "failed to begin transaction to set arrival time for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.String(), station)
		return
	}

	_, err = tx.Exec(`
UPDATE arrivals
SET arrival_time = $1
WHERE arrivals.identifier = $2
  AND arrival_time IS NULL`,
		arrivalTime,
		ArrivalIdentifierFor(dir, line, trainID, runFirstEventMoment, station),
	)
	if err != nil {
		rollback(tx, a.Logger)
		err = errors.Wrapf(err, "failed to set arrival time for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.String(), station)
		return
	}

	err = touchRun(tx, dir, line, trainID, runFirstEventMoment, eventTime)
	if err != nil {
		rollback(tx, a.Logger)
		return
	}

	err = errors.Wrapf(tx.Commit(), "failed to commit transaction when setting arrival time for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.String(), station)
	return
}

func touchRun(tx *sql.Tx, dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, touchMoment EasternTime) (err error) {
	res, err := tx.Exec(`
UPDATE runs
SET most_recent_event_moment = $1
WHERE identifier = $2`,
		touchMoment,
		RunIdentifierFor(dir, line, trainID, runFirstEventMoment),
	)
	if err != nil {
		err = errors.Wrapf(err, "failed to touch run for dir `%s` line `%s` train `%s` first event moment `%s`", dir, line, trainID, runFirstEventMoment.String())
		return
	}

	i, err := res.RowsAffected()
	if err != nil {
		err = errors.Wrapf(err, "received malformed result when touching run for dir `%s` line `%s` train `%s` first event moment `%s`", dir, line, trainID, runFirstEventMoment.String())
		return
	}
	if i != 1 {
		err = fmt.Errorf("touch-run query unexpectedly affected %v rows - expected 1", i)
		return
	}

	return
}

func rollback(tx *sql.Tx, logger *zap.Logger) {
	if err := tx.Rollback(); err != nil {
		logger.Error(fmt.Sprintf("failed rolling back transaction: %s", err.Error()))
	}
}
