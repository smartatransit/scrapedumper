package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/smartatransit/scrapedumper/pkg/martaapi"
	"go.uber.org/zap"
)

type LastestEstimate struct {
	Direction   string
	Line        string
	Station     string
	DirectionID *uint
	LineID      *uint
	StationID   *uint
	TrainID     string
	Destination string

	NextArrival EasternTime
	EventTime   EasternTime
}

//Repository implements interactions with Postgres through GORM
//go:generate counterfeiter . Repository
type Repository interface {
	EnsureTables(thirdRail bool) error

	GetLatestRunStartMomentFor(dir martaapi.Direction, line martaapi.Line, trainID string, asOfMoment EasternTime) (runFirstEventMoment EasternTime, mostRecentEventTime EasternTime, err error)
	CreateRunRecord(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, correctedLine martaapi.Line, correctedDirection martaapi.Direction, lineID *uint, dirID *uint) (err error)
	EnsureArrivalRecord(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, station martaapi.Station, stationID *uint) (err error)
	AddArrivalEstimate(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, station martaapi.Station, eventTime EasternTime, estimate EasternTime) (err error)
	SetArrivalTime(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, station martaapi.Station, eventTime EasternTime, arrival EasternTime) (err error)

	GetRecentlyActiveRuns(touchThreshold EasternTime) (runs map[string]Run, err error)
	GetLatestEstimates(stationID uint) (res []LastestEstimate, err error)

	DeleteStaleRuns(threshold EasternTime) (estimatesDropped int64, arrivalsDropped int64, runsDropped int64, err error)
}

//NewRepository creates a new postgres respository
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

//ParseRunGroupIdentifier converts a run group identifier back into metadata
func ParseRunGroupIdentifier(id string) (dir martaapi.Direction, line martaapi.Line, trainID string) {
	parts := strings.Split(id, "_")
	return martaapi.Direction(parts[0]), martaapi.Line(parts[1]), parts[2]
}

//EnsureTables ensures that all necessary tables exist. Three fields (line_id,
//station_id, and direction_id) are always included, however if thirdRail is
//false, they are always left empty. Including them in both cases simplifies
//our update/select queries.
func (a *RepositoryAgent) EnsureTables(thirdRail bool) error {
	runsExtras := `
	line_id integer,
	direction_id integer,
`
	if thirdRail {
		runsExtras = `
	line_id integer REFERENCES lines(id),
	direction_id integer REFERENCES directions(id),
`
	}

	_, err := a.DB.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS runs
(	identifier varchar,
	run_group_identifier varchar NOT NULL,
	corrected_line varchar NOT NULL,
	corrected_direction varchar NOT NULL,
	most_recent_event_moment varchar NOT NULL,
	run_first_event_moment varchar NOT NULL,%s

	PRIMARY KEY (identifier)
)`, runsExtras))
	if err != nil {
		return errors.Wrapf(err, "failed to ensure runs table")
	}

	arrivalExtras := `
	station_id integer,
`
	if thirdRail {
		arrivalExtras = `
	station_id integer REFERENCES stations(id),
`
	}

	_, err = a.DB.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS arrivals
(	identifier varchar,
	run_identifier varchar NOT NULL,
	station varchar NOT NULL,
	arrival_time varchar,%s

	PRIMARY KEY (identifier)
)`, arrivalExtras))
	if err != nil {
		return errors.Wrapf(err, "failed to ensure arrivals table")
	}

	_, err = a.DB.Exec(`
CREATE TABLE IF NOT EXISTS estimates
(	identifier varchar,
	run_identifier varchar NOT NULL,
	arrival_identifier varchar NOT NULL,
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
func (a *RepositoryAgent) CreateRunRecord(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, correctedLine martaapi.Line, correctedDirection martaapi.Direction, lineID *uint, dirID *uint) (err error) {
	res, err := a.DB.Exec(`
INSERT INTO runs
(identifier, run_group_identifier, most_recent_event_moment, run_first_event_moment, corrected_line, corrected_direction, line_id, direction_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		RunIdentifierFor(dir, line, trainID, runFirstEventMoment),
		RunGroupIdentifierFor(dir, line, trainID),
		runFirstEventMoment, //most_recent_event_moment
		runFirstEventMoment,
		correctedLine,
		correctedDirection,
		lineID,
		dirID,
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
func (a *RepositoryAgent) EnsureArrivalRecord(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, station martaapi.Station, stationID *uint) (err error) {
	_, err = a.DB.Exec(`
INSERT INTO arrivals
(identifier, run_identifier, station, station_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING`,
		ArrivalIdentifierFor(dir, line, trainID, runFirstEventMoment, station),
		RunIdentifierFor(dir, line, trainID, runFirstEventMoment),
		station,
		stationID,
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

func (a *RepositoryAgent) DeleteStaleRuns(threshold EasternTime) (estimatesDropped int64, arrivalsDropped int64, runsDropped int64, err error) {
	tx, err := a.DB.Begin()
	if err != nil {
		err = errors.Wrap(err, "failed to begin transaction to delete stale runs")
		return
	}

	res, err := tx.Exec(`
DELETE FROM estimates
USING runs
WHERE runs.identifier = estimates.run_identifier
	AND runs.most_recent_event_moment < $1`,
		threshold,
	)
	if err != nil {
		rollback(tx, a.Logger)
		err = errors.Wrap(err, "failed to drop estimates for stale runs")
		return
	}
	if estimatesDropped, err = res.RowsAffected(); err != nil {
		rollback(tx, a.Logger)
		err = errors.Wrap(err, "received malformed result when dropping stale estimates")
		return
	}

	res, err = tx.Exec(`
DELETE FROM arrivals
USING runs
WHERE runs.identifier = arrivals.run_identifier
	AND runs.most_recent_event_moment < $1`,
		threshold,
	)
	if err != nil {
		rollback(tx, a.Logger)
		err = errors.Wrap(err, "failed to drop arrivals for stale runs")
		return
	}
	if arrivalsDropped, err = res.RowsAffected(); err != nil {
		rollback(tx, a.Logger)
		err = errors.Wrap(err, "received malformed result when dropping stale arrivals")
		return
	}

	res, err = tx.Exec(`
DELETE FROM runs
WHERE most_recent_event_moment < $1`,
		threshold,
	)
	if err != nil {
		rollback(tx, a.Logger)
		err = errors.Wrap(err, "failed to drop stale runs")
		return
	}
	if runsDropped, err = res.RowsAffected(); err != nil {
		rollback(tx, a.Logger)
		err = errors.Wrap(err, "received malformed result when dropping stale runs")
		return
	}

	err = errors.Wrapf(tx.Commit(), "failed to commit transaction when dropping stale runs")
	return
}

//GetRecentlyActiveRuns collects all the data about any runs that have been updated
//since touchThreshold. The Run#Finished method can be used to determine which runs
//have arrived at their terminal station, and can therefore be removed from state.
func (a *RepositoryAgent) GetRecentlyActiveRuns(touchThreshold EasternTime) (runs map[string]Run, err error) {
	rows, err := a.DB.Query(`
SELECT runs.identifier, runs.run_group_identifier,
  runs.corrected_line, runs.corrected_direction,
  runs.most_recent_event_moment, runs.run_first_event_moment,
  arrivals.identifier, arrivals.station, arrivals.arrival_time,
  estimates.estimate_moment, estimates.estimated_arrival_time

FROM runs
JOIN arrivals
  ON runs.identifier = arrivals.run_identifier
JOIN estimates
  ON arrivals.identifier = estimates.arrival_identifier

WHERE runs.most_recent_event_moment > $1
ORDER BY estimates.identifier ASC`,
		touchThreshold,
	)
	if err != nil {
		err = errors.Wrapf(err, "failed to get active runs")
		return
	}

	runs = map[string]Run{}
	for rows.Next() {
		var run Run
		var arrival Arrival
		var estimateMoment, estimatedArrivalTime EasternTime
		err = rows.Scan(
			&run.Identifier,
			&run.RunGroupIdentifier,
			&run.CorrectedLine,
			&run.CorrectedDirection,
			&run.MostRecentEventMoment,
			&run.RunFirstEventMoment,
			&arrival.Identifier,
			&arrival.Station,
			&arrival.ArrivalTime,
			&estimateMoment,
			&estimatedArrivalTime,
		)
		if err != nil {
			err = errors.Wrapf(err, "failed to scan run")
			return
		}

		if seenRun, ok := runs[run.Identifier]; ok {
			run = seenRun
		} else {
			run.Arrivals = map[martaapi.Station]Arrival{}
			runs[run.Identifier] = run
		}

		if seenArrival, ok := run.Arrivals[arrival.Station]; ok {
			//if we've already seen this run, use the existing copy
			arrival = seenArrival
		} else {
			arrival.Estimates = map[EasternTime]EasternTime{}
			run.Arrivals[arrival.Station] = arrival
		}

		arrival.Estimates[estimateMoment] = estimatedArrivalTime
	}

	return
}

//GetRecentlyActiveRuns collects all the data about any runs that have been updated
//since touchThreshold. The Run#Finished method can be used to determine which runs
//have arrived at their terminal station, and can therefore be removed from state.
func (a *RepositoryAgent) GetLatestEstimates(stationID uint) (res []LastestEstimate, err error) {
	rows, err := a.DB.Query(`
WITH station_estimates AS (
  SELECT runs.run_group_identifier,
    estimates.estimated_arrival_time,
    estimates.estimate_moment,

    directions.name,
    lines.name,
    stations.name,

    directions.id,
    lines.id,
    stations.id,

    ROW_NUMBER() OVER (PARTITION BY runs.run_group_identifier
      ORDER BY estimates.estimate_moment DESC) AS rank
  FROM arrivals
  JOIN estimates
    ON estimates.arrival_identifier = arrivals.identifier
  JOIN runs
    ON arrivals.run_identifier = runs.identifier

  JOIN lines
    ON lines.id = runs.line_id
  JOIN directions
    ON directions.id = runs.direction_id
  JOIN stations
    ON stations.id = arrivals.station_id
  WHERE arrivals.station_id = $1
    AND arrivals.arrival_time IS NULL)

SELECT *
  FROM station_estimates
  WHERE rank = 1
`,
		stationID,
	)
	if err != nil {
		err = errors.Wrapf(err, "failed to fetch arrival estimates")
		return
	}

	res = []LastestEstimate{}
	for rows.Next() {
		var sched LastestEstimate
		var rgIdentifier, dir, line, station string
		var dirID, lineID, stationID sql.NullInt32
		err = rows.Scan(
			&rgIdentifier,
			&sched.NextArrival,
			&sched.EventTime,

			&dir,
			&line,
			&station,
			&dirID,
			&lineID,
			&stationID,
		)
		if err != nil {
			err = errors.Wrapf(err, "failed to scan run")
			return
		}

		_, _, train := ParseRunGroupIdentifier(rgIdentifier)
		sched.Direction = dir
		sched.Line = line
		sched.Station = station
		sched.TrainID = train
		sched.Destination = string(martaapi.Termini[martaapi.Line(line)][martaapi.Direction(dir)])
		if dirID.Valid {
			var u uint = uint(dirID.Int32)
			sched.DirectionID = &u
		}
		if lineID.Valid {
			var u uint = uint(lineID.Int32)
			sched.LineID = &u
		}
		if stationID.Valid {
			var u uint = uint(stationID.Int32)
			sched.StationID = &u
		}

		res = append(res, sched)
	}

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
