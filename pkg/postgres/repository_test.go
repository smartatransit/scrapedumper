package postgres_test

import (
	"database/sql"
	"errors"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/bipol/scrapedumper/pkg/postgres"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func easternDate(year int, month time.Month, day, hour, min, sec, nsec int) postgres.EasternTime {
	return postgres.EasternTime(time.Date(year, month, day, hour, min, sec, nsec, postgres.EasternTimeZone))
}

var _ = Describe("Repository", func() {
	var (
		db    *sql.DB
		smock sqlmock.Sqlmock

		repo postgres.Repository
	)

	BeforeEach(func() {
		var err error
		db, smock, err = sqlmock.New()
		Expect(err).To(BeNil())
	})

	JustBeforeEach(func() {
		repo = postgres.NewRepository(zap.NewNop(), db)
	})

	Describe("EnsureTables", func() {
		var callErr error

		var expectRunsTableExec = func() *sqlmock.ExpectedExec {
			return smock.ExpectExec(`
CREATE TABLE IF NOT EXISTS runs
\(	identifier varchar,
	run_group_identifier varchar NOT NULL,
	corrected_line varchar NOT NULL,
	corrected_direction varchar NOT NULL,
	most_recent_event_moment varchar NOT NULL,
	run_first_event_moment varchar NOT NULL,
	PRIMARY KEY \(identifier\)
\)`)
		}
		var expectArrivalsTableExec = func() *sqlmock.ExpectedExec {
			return smock.ExpectExec(`
CREATE TABLE IF NOT EXISTS arrivals
\(	identifier varchar,
	run_identifier varchar NOT NULL,
	station varchar NOT NULL,
	arrival_time varchar,
	PRIMARY KEY \(identifier\)
\)`)
		}
		var expectEstimatesTableExec = func() *sqlmock.ExpectedExec {
			return smock.ExpectExec(`
CREATE TABLE IF NOT EXISTS estimates
\(	identifier varchar,
	run_identifier varchar NOT NULL,
	arrival_identifier varchar NOT NULL,
	estimate_moment varchar NOT NULL,
	estimated_arrival_time varchar NOT NULL,
	PRIMARY KEY \(identifier\)
\)`)
		}
		var expectRunGroupIndexExec = func() *sqlmock.ExpectedExec {
			return smock.ExpectExec(`CREATE INDEX ON runs USING btree\(run_group_identifier\)`)
		}
		var expectRunIndexExec = func() *sqlmock.ExpectedExec {
			return smock.ExpectExec(`CREATE INDEX ON arrivals USING btree\(run_identifier\)`)
		}
		var expectArrivalIndexExec = func() *sqlmock.ExpectedExec {
			return smock.ExpectExec(`CREATE INDEX ON estimates USING btree\(arrival_identifier\)`)
		}
		var expectEstimatesByRunIndexExec = func() *sqlmock.ExpectedExec {
			return smock.ExpectExec(`CREATE INDEX ON estimates USING btree\(run_identifier\)`)
		}
		var expectLatestRunIndexExec = func() *sqlmock.ExpectedExec {
			return smock.ExpectExec(`
CREATE INDEX ON runs USING btree\(
	run_group_identifier,
	run_first_event_moment DESC,
	most_recent_event_moment DESC
\)`)
		}

		JustBeforeEach(func() {
			callErr = repo.EnsureTables()
		})
		When("the runs table fails", func() {
			BeforeEach(func() {
				expectRunsTableExec().WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to ensure runs table: exec failed"))
			})
		})
		When("the arrivals table fails", func() {
			BeforeEach(func() {
				expectRunsTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectArrivalsTableExec().WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to ensure arrivals table: exec failed"))
			})
		})
		When("the estimates table fails", func() {
			BeforeEach(func() {
				expectRunsTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectArrivalsTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectEstimatesTableExec().WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to ensure estimates table: exec failed"))
			})
		})
		When("the run group index fails", func() {
			BeforeEach(func() {
				expectRunsTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectArrivalsTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectEstimatesTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectRunGroupIndexExec().WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to index runs by run group: exec failed"))
			})
		})
		When("the run index fails", func() {
			BeforeEach(func() {
				expectRunsTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectArrivalsTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectEstimatesTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectRunGroupIndexExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectRunIndexExec().WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to index arrivals by run: exec failed"))
			})
		})
		When("the arrival index fails", func() {
			BeforeEach(func() {
				expectRunsTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectArrivalsTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectEstimatesTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectRunGroupIndexExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectRunIndexExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectArrivalIndexExec().WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to index estimates by arrival: exec failed"))
			})
		})
		When("the estimates-by-run index fails", func() {
			BeforeEach(func() {
				expectRunsTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectArrivalsTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectEstimatesTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectRunGroupIndexExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectRunIndexExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectArrivalIndexExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectEstimatesByRunIndexExec().WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to index estimates by run: exec failed"))
			})
		})
		When("the latest-matching-run index fails", func() {
			BeforeEach(func() {
				expectRunsTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectArrivalsTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectEstimatesTableExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectRunGroupIndexExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectRunIndexExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectArrivalIndexExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectEstimatesByRunIndexExec().WillReturnResult(sqlmock.NewResult(0, 0))
				expectLatestRunIndexExec().WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to index runs for upserting: exec failed"))
			})
		})
	})

	Describe("GetLatestRunStartMomentFor", func() {
		var (
			runFirstEventMoment postgres.EasternTime
			mostRecentEventTime postgres.EasternTime
			callErr             error

			query *sqlmock.ExpectedQuery
			rows  *sqlmock.Rows
		)
		BeforeEach(func() {
			query = smock.ExpectQuery(`
SELECT run_first_event_moment, runs.most_recent_event_moment
FROM arrivals JOIN runs ON runs.identifier = arrivals.run_identifier
WHERE run_group_identifier = \$1 AND runs.most_recent_event_moment <= \$2
ORDER BY run_first_event_moment DESC, runs.most_recent_event_moment DESC, arrivals.identifier ASC
LIMIT 1`).
				WithArgs("N_GOLD_193230", "2019-08-05T18:15:16-04:00")

			rows = sqlmock.NewRows([]string{"run_first_event_moment", "most_recent_event_moment"})
			query.WillReturnRows(rows)
		})
		JustBeforeEach(func() {
			runFirstEventMoment, mostRecentEventTime, callErr = repo.GetLatestRunStartMomentFor(
				martaapi.Direction("N"),
				martaapi.Line("GOLD"),
				"193230",
				easternDate(2019, time.August, 5, 18, 15, 16, 0),
			)
		})
		When("the query fails", func() {
			BeforeEach(func() {
				query.WillReturnError(errors.New("query failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to query latest run start moment for dir `N` line `GOLD` and train `193230`: query failed"))
			})
		})
		//this provides test coverage to the EasternTime#Scan method
		When("the query returns an int for the timestamps", func() {
			BeforeEach(func() {
				rows := sqlmock.NewRows([]string{"run_first_event_moment", "most_recent_event_moment"})
				rows.AddRow(5, 5)
				query.WillReturnRows(rows)
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to query latest run start moment for dir `N` line `GOLD` and train `193230`: sql: Scan error on column index 0, name \"run_first_event_moment\": expected string, got int64"))
			})
		})
		When("no record is found", func() {
			It("returns zero times", func() {
				Expect(runFirstEventMoment).To(BeZero())
				Expect(mostRecentEventTime).To(BeZero())
				Expect(callErr).To(BeNil())
			})
		})
		When("all goes well", func() {
			BeforeEach(func() {
				rows.AddRow(
					easternDate(2019, time.August, 5, 18, 15, 16, 0),
					easternDate(2019, time.August, 5, 18, 34, 16, 0),
				)
			})
			It("succeeds", func() {
				Expect(callErr).To(BeNil())
				Expect(runFirstEventMoment).To(Equal(easternDate(2019, time.August, 5, 18, 15, 16, 0)))
				Expect(mostRecentEventTime).To(Equal(easternDate(2019, time.August, 5, 18, 34, 16, 0)))
			})
		})
	})

	Describe("CreateRunRecord", func() {
		var (
			callErr error

			exec *sqlmock.ExpectedExec
		)
		BeforeEach(func() {
			exec = smock.ExpectExec(`
INSERT INTO runs
\(identifier, run_group_identifier, most_recent_event_moment, run_first_event_moment, corrected_line, corrected_direction\)
VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`).
				WithArgs(
					"N_GOLD_193230_2019-08-05T18:15:16-04:00",
					"N_GOLD_193230",
					easternDate(2019, time.August, 5, 18, 15, 16, 0),
					easternDate(2019, time.August, 5, 18, 15, 16, 0),
					"RED",
					"S",
				)
			exec.WillReturnResult(sqlmock.NewResult(0, 1))
		})
		JustBeforeEach(func() {
			callErr = repo.CreateRunRecord(
				martaapi.Direction("N"),
				martaapi.Line("GOLD"),
				"193230",
				easternDate(2019, time.August, 5, 18, 15, 16, 0),
				martaapi.Line("RED"),
				martaapi.Direction("S"),
			)
		})
		When("the query fails", func() {
			BeforeEach(func() {
				exec.WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to create run for dir `N` line `GOLD` train `193230` and first event moment `2019-08-05T18:15:16-04:00`: exec failed"))
			})
		})
		When("the update result errors out", func() {
			BeforeEach(func() {
				exec.WillReturnResult(sqlmock.NewErrorResult(errors.New("can't get rows-affected")))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("received malformed result when creating run for dir `N` line `GOLD` train `193230` and first event moment `2019-08-05T18:15:16-04:00`: can't get rows-affected"))
			})
		})
		When("the update doesn't affect any rows", func() {
			BeforeEach(func() {
				exec.WillReturnResult(sqlmock.NewResult(0, 0))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("create-run query unexpectedly affected 0 rows - expected 1"))
			})
		})
		When("all goes well", func() {
			It("fails", func() {
				Expect(callErr).To(BeNil())
			})
		})
	})

	Describe("EnsureArrivalRecord", func() {
		var (
			callErr error

			exec *sqlmock.ExpectedExec
		)
		BeforeEach(func() {
			exec = smock.ExpectExec(`
INSERT INTO arrivals
\(identifier, run_identifier, station\)
VALUES \(\$1, \$2, \$3\)
ON CONFLICT DO NOTHING`).
				WithArgs(
					"N_GOLD_193230_2019-08-05T18:15:16-04:00_FIVE POINTS",
					"N_GOLD_193230_2019-08-05T18:15:16-04:00",
					"FIVE POINTS",
				)
			exec.WillReturnResult(sqlmock.NewResult(0, 0))
		})
		JustBeforeEach(func() {
			callErr = repo.EnsureArrivalRecord(
				martaapi.Direction("N"),
				martaapi.Line("GOLD"),
				"193230",
				easternDate(2019, time.August, 5, 18, 15, 16, 0),
				martaapi.Station("FIVE POINTS"),
			)
		})
		When("the query fails", func() {
			BeforeEach(func() {
				exec.WillReturnError(errors.New("query failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to ensure arrival for dir `N` line `GOLD` train `193230` first event moment `2019-08-05T18:15:16-04:00` and station `FIVE POINTS`: query failed"))
			})
		})
	})

	Describe("AddArrivalEstimate", func() {
		var (
			callErr error

			begin     *sqlmock.ExpectedBegin
			firstExec *sqlmock.ExpectedExec
			touchExec *sqlmock.ExpectedExec
			eventTime postgres.EasternTime
		)
		BeforeEach(func() {
			begin = smock.ExpectBegin()

			firstExec = smock.ExpectExec(`
INSERT INTO estimates
\(identifier, run_identifier, arrival_identifier, estimate_moment, estimated_arrival_time\)
VALUES \(\$1, \$2, \$3, \$4, \$5\)
ON CONFLICT DO NOTHING`)
			firstExec.WillReturnResult(sqlmock.NewResult(0, 0))

			touchExec = smock.ExpectExec(`
UPDATE runs
SET most_recent_event_moment = \$1
WHERE identifier = \$2`)

			eventTime = easternDate(2019, time.August, 5, 20, 15, 16, 0)
			touchExec.WillReturnResult(sqlmock.NewResult(0, 1))
		})
		JustBeforeEach(func() {
			firstExec.WithArgs(
				"N_GOLD_193230_2019-08-05T18:15:16-04:00_FIVE POINTS_2019-08-05T20:15:16-04:00",
				"N_GOLD_193230_2019-08-05T18:15:16-04:00",
				"N_GOLD_193230_2019-08-05T18:15:16-04:00_FIVE POINTS",
				easternDate(2019, time.August, 5, 20, 15, 16, 0),
				easternDate(2019, time.August, 5, 22, 15, 16, 0),
			)

			touchExec.WithArgs(
				easternDate(2019, time.August, 5, 20, 15, 16, 0),
				"N_GOLD_193230_2019-08-05T18:15:16-04:00",
			)

			callErr = repo.AddArrivalEstimate(
				martaapi.Direction("N"),
				martaapi.Line("GOLD"),
				"193230",
				easternDate(2019, time.August, 5, 18, 15, 16, 0),
				martaapi.Station("FIVE POINTS"),
				eventTime,
				easternDate(2019, time.August, 5, 22, 15, 16, 0),
			)
		})
		When("beginning the transaction fails", func() {
			BeforeEach(func() {
				begin.WillReturnError(errors.New("begin failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to begin transaction to add arrival estimate for dir `N` line `GOLD` train `193230` first event moment `2019-08-05T18:15:16-04:00` and station `FIVE POINTS`: begin failed"))
			})
		})
		When("the first update fails", func() {
			BeforeEach(func() {
				firstExec.WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to add arrival estimate for dir `N` line `GOLD` train `193230` first event moment `2019-08-05T18:15:16-04:00` and station `FIVE POINTS`: exec failed"))
			})
		})
		When("the touch update fails", func() {
			BeforeEach(func() {
				touchExec.WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to touch run for dir `N` line `GOLD` train `193230` first event moment `2019-08-05T18:15:16-04:00`: exec failed"))
			})
		})
		When("the touch result errors out", func() {
			BeforeEach(func() {
				touchExec.WillReturnResult(sqlmock.NewErrorResult(errors.New("can't get rows-affected")))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("received malformed result when touching run for dir `N` line `GOLD` train `193230` first event moment `2019-08-05T18:15:16-04:00`: can't get rows-affected"))
			})
		})
		When("the touch doesn't affect any rows", func() {
			BeforeEach(func() {
				touchExec.WillReturnResult(sqlmock.NewResult(0, 0))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("touch-run query unexpectedly affected 0 rows - expected 1"))
			})
		})
		When("the update succeeds", func() {
			BeforeEach(func() {
				smock.ExpectCommit().WillReturnError(errors.New("commit failed"))
			})
			It("succeeds", func() {
				Expect(callErr).To(MatchError("failed to commit transaction when setting arrival time for dir `N` line `GOLD` train `193230` first event moment `2019-08-05T18:15:16-04:00` and station `FIVE POINTS`: commit failed"))
			})
		})
	})

	Describe("SetArrivalTime", func() {
		var (
			callErr error

			begin     *sqlmock.ExpectedBegin
			firstExec *sqlmock.ExpectedExec
			touchExec *sqlmock.ExpectedExec
		)
		BeforeEach(func() {
			begin = smock.ExpectBegin()

			firstExec = smock.ExpectExec(`
UPDATE arrivals
SET arrival_time = \$1
WHERE arrivals.identifier = \$2
  AND arrival_time IS NULL`).
				WithArgs(
					easternDate(2019, time.August, 5, 22, 15, 16, 0),
					"N_GOLD_193230_2019-08-05T18:15:16-04:00_FIVE POINTS",
				)
			firstExec.WillReturnResult(sqlmock.NewResult(0, 1))

			touchExec = smock.ExpectExec(`
UPDATE runs
SET most_recent_event_moment = \$1
WHERE identifier = \$2`).
				WithArgs(
					easternDate(2019, time.August, 5, 20, 15, 16, 0),
					"N_GOLD_193230_2019-08-05T18:15:16-04:00",
				)
			touchExec.WillReturnResult(sqlmock.NewResult(0, 1))
		})
		JustBeforeEach(func() {
			callErr = repo.SetArrivalTime(
				martaapi.Direction("N"),
				martaapi.Line("GOLD"),
				"193230",
				easternDate(2019, time.August, 5, 18, 15, 16, 0),
				martaapi.Station("FIVE POINTS"),
				easternDate(2019, time.August, 5, 20, 15, 16, 0),
				easternDate(2019, time.August, 5, 22, 15, 16, 0),
			)
		})
		When("beginning the transaction fails", func() {
			BeforeEach(func() {
				begin.WillReturnError(errors.New("begin failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to begin transaction to set arrival time for dir `N` line `GOLD` train `193230` first event moment `2019-08-05T18:15:16-04:00` and station `FIVE POINTS`: begin failed"))
			})
		})
		When("the update fails", func() {
			BeforeEach(func() {
				firstExec.WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to set arrival time for dir `N` line `GOLD` train `193230` first event moment `2019-08-05T18:15:16-04:00` and station `FIVE POINTS`: exec failed"))
			})
		})
		When("touching the run fails", func() {
			BeforeEach(func() {
				touchExec.WillReturnResult(sqlmock.NewResult(0, 0))
			})
			It("succeeds", func() {
				Expect(callErr).To(MatchError("touch-run query unexpectedly affected 0 rows - expected 1"))
			})
		})
		When("committing the transaction fails", func() {
			BeforeEach(func() {
				smock.ExpectCommit().WillReturnError(errors.New("commit failed"))
			})
			It("succeeds", func() {
				Expect(callErr).To(MatchError("failed to commit transaction when setting arrival time for dir `N` line `GOLD` train `193230` first event moment `2019-08-05T18:15:16-04:00` and station `FIVE POINTS`: commit failed"))
			})
		})
	})

	Describe("DeleteStaleRuns", func() {
		var (
			callErr error

			begin         *sqlmock.ExpectedBegin
			estimatesExec *sqlmock.ExpectedExec
			arrivalsExec  *sqlmock.ExpectedExec
			runsExec      *sqlmock.ExpectedExec
		)
		BeforeEach(func() {
			begin = smock.ExpectBegin()

			estimatesExec = smock.ExpectExec(`
DELETE FROM estimates
USING runs
WHERE runs.identifier = estimates.run_identifier
	AND runs.most_recent_event_moment < \$1`).
				WithArgs(
					easternDate(2019, time.August, 5, 22, 15, 16, 0),
				)
			estimatesExec.WillReturnResult(sqlmock.NewResult(0, 1))

			arrivalsExec = smock.ExpectExec(`
DELETE FROM arrivals
USING runs
WHERE runs.identifier = arrivals.run_identifier
	AND runs.most_recent_event_moment < \$1`).
				WithArgs(
					easternDate(2019, time.August, 5, 22, 15, 16, 0),
				)
			arrivalsExec.WillReturnResult(sqlmock.NewResult(0, 1))

			runsExec = smock.ExpectExec(`
DELETE FROM runs
WHERE most_recent_event_moment < \$1`).
				WithArgs(
					easternDate(2019, time.August, 5, 22, 15, 16, 0),
				)
			runsExec.WillReturnResult(sqlmock.NewResult(0, 1))
		})
		JustBeforeEach(func() {
			callErr = repo.DeleteStaleRuns(easternDate(2019, time.August, 5, 22, 15, 16, 0))
		})
		When("beginning the transaction fails", func() {
			BeforeEach(func() {
				begin.WillReturnError(errors.New("begin failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to begin transaction to delete stale runs: begin failed"))
			})
		})
		When("the update fails", func() {
			BeforeEach(func() {
				estimatesExec.WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to drop estimates for stale runs: exec failed"))
			})
		})
		When("the update fails", func() {
			BeforeEach(func() {
				arrivalsExec.WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to drop arrivals for stale runs: exec failed"))
			})
		})
		When("the update fails", func() {
			BeforeEach(func() {
				runsExec.WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to drop stale runs: exec failed"))
			})
		})
		When("committing the transaction fails", func() {
			BeforeEach(func() {
				smock.ExpectCommit().WillReturnError(errors.New("commit failed"))
			})
			It("succeeds", func() {
				Expect(callErr).To(MatchError("failed to commit transaction when dropping stale runs: commit failed"))
				Expect(smock.ExpectationsWereMet()).To(BeNil())
			})
		})
	})
})
