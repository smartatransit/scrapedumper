package postgres_test

import (
	"database/sql"
	"errors"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/bipol/scrapedumper/pkg/marta"
	"github.com/bipol/scrapedumper/pkg/postgres"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dump", func() {
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
		repo = postgres.NewRepository(db)
	})

	Context("EnsureTables", func() {
		var (
			callErr error

			exec *sqlmock.ExpectedExec
		)
		BeforeEach(func() {
			exec = smock.ExpectExec(`
CREATE TABLE IF NOT EXISTS "arrivals"
\(	"identifier" text,
	"run_identifier" text,
	"run_group_identifier" text,
	"most_recent_event_time" timestamp,
	"direction" text,
	"line" text,
	"train_id" text,
	"run_first_event_moment" timestamp,
	"station" text,
	"arrival_time" timestamp,
	"arrival_estimates" jsonb,
	PRIMARY KEY \("identifier"\)
\)`)
		})
		JustBeforeEach(func() {
			callErr = repo.EnsureTables()
		})
		When("the query fails", func() {
			BeforeEach(func() {
				exec.WillReturnError(errors.New("query failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to ensure arrivals table: query failed"))
			})
		})
	})

	Context("GetLatestRunStartMomentFor", func() {
		var (
			runFirstEventMoment time.Time
			mostRecentEventTime time.Time
			callErr             error

			query *sqlmock.ExpectedQuery
			rows  *sqlmock.Rows
		)
		BeforeEach(func() {
			query = smock.ExpectQuery(`
SELECT run_first_event_moment, most_recent_event_time
FROM "arrivals"
WHERE run_group_identifier = \$1
ORDER BY run_first_event_moment DESC, most_recent_event_time DESC, "arrivals"."identifier" ASC
LIMIT 1`).
				WithArgs("N_GOLD_193230")

			rows = sqlmock.NewRows([]string{"run_first_event_moment", "most_recent_event_time"})
			query.WillReturnRows(rows)
		})
		JustBeforeEach(func() {
			runFirstEventMoment, mostRecentEventTime, callErr = repo.GetLatestRunStartMomentFor(
				marta.Direction("N"),
				marta.Line("GOLD"),
				"193230",
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
					time.Date(2019, time.August, 5, 18, 15, 16, 0, postgres.EasternTime),
					time.Date(2019, time.August, 5, 18, 34, 16, 0, postgres.EasternTime),
				)
			})
			It("succeeds", func() {
				Expect(runFirstEventMoment).To(Equal(time.Date(2019, time.August, 5, 18, 15, 16, 0, postgres.EasternTime)))
				Expect(mostRecentEventTime).To(Equal(time.Date(2019, time.August, 5, 18, 34, 16, 0, postgres.EasternTime)))
				Expect(callErr).To(BeNil())
			})
		})
	})

	Context("EnsureArrivalRecord", func() {
		var (
			callErr error

			exec *sqlmock.ExpectedExec
		)
		BeforeEach(func() {
			exec = smock.ExpectExec(`
INSERT INTO "arrivals"
\("identifier", "run_identifier", "run_group_identifier", "most_recent_event_time", "direction", "line", "train_id", "run_first_event_moment", "station", "arrival_time", "arrival_estimates"\)
VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, \$10, \$11\)
ON CONFLICT DO NOTHING`).
				WithArgs(
					"N_GOLD_193230_2019-08-05T18:15:16-04:00_FIVE POINTS",
					"N_GOLD_193230_2019-08-05T18:15:16-04:00",
					"N_GOLD_193230",
					time.Date(2019, time.August, 5, 18, 15, 16, 0, postgres.EasternTime),
					"N",
					"GOLD",
					"193230",
					time.Date(2019, time.August, 5, 18, 15, 16, 0, postgres.EasternTime),
					"FIVE POINTS",
					time.Time{},
					postgres.ArrivalEstimates(map[time.Time]time.Time{}),
				)
			exec.WillReturnResult(sqlmock.NewResult(0, 0))
		})
		JustBeforeEach(func() {
			callErr = repo.EnsureArrivalRecord(
				marta.Direction("N"),
				marta.Line("GOLD"),
				"193230",
				time.Date(2019, time.August, 5, 18, 15, 16, 0, postgres.EasternTime),
				marta.Station("FIVE POINTS"),
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
		When("all goes well", func() {
			It("succeeds", func() {
				Expect(callErr).To(BeNil())
			})
		})
	})

	Context("AddArrivalEstimate", func() {
		var (
			callErr error

			exec *sqlmock.ExpectedExec
		)
		BeforeEach(func() {
			exec = smock.ExpectExec(`
UPDATE "arrivals"
SET \("arrival_estimates", "most_recent_event_time"\)
  = \("arrival_estimates" || \$1, \$2\)
WHERE "arrivals"."identifier" = \$3`).
				WithArgs(
					`{"2019-08-05T20:15:16-04:00":"2019-08-05T22:15:16-04:00"}`,
					time.Date(2019, time.August, 5, 20, 15, 16, 0, postgres.EasternTime),
					"N_GOLD_193230_2019-08-05T18:15:16-04:00_FIVE POINTS",
				)
			exec.WillReturnResult(sqlmock.NewResult(0, 0))
		})
		JustBeforeEach(func() {
			callErr = repo.AddArrivalEstimate(
				marta.Direction("N"),
				marta.Line("GOLD"),
				"193230",
				time.Date(2019, time.August, 5, 18, 15, 16, 0, postgres.EasternTime),
				marta.Station("FIVE POINTS"),
				time.Date(2019, time.August, 5, 20, 15, 16, 0, postgres.EasternTime),
				time.Date(2019, time.August, 5, 22, 15, 16, 0, postgres.EasternTime),
			)
		})
		When("the query fails", func() {
			BeforeEach(func() {
				exec.WillReturnError(errors.New("query failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to add arrival estimate for dir `N` line `GOLD` train `193230` first event moment `2019-08-05T18:15:16-04:00` and station `FIVE POINTS`: query failed"))
			})
		})
		When("all goes well", func() {
			It("succeeds", func() {
				Expect(callErr).To(BeNil())
			})
		})
	})

	Context("SetArrivalTime", func() {
		var (
			callErr error

			exec *sqlmock.ExpectedExec
		)
		BeforeEach(func() {
			exec = smock.ExpectExec(`
UPDATE "arrivals"
SET \("arrival_time", "most_recent_event_time"\)
  = \(\$1, \$2\)
WHERE "arrivals"."identifier" = \$3
  AND "arrival_time" = \$4`).
				WithArgs(
					time.Date(2019, time.August, 5, 22, 15, 16, 0, postgres.EasternTime),
					time.Date(2019, time.August, 5, 20, 15, 16, 0, postgres.EasternTime),
					"N_GOLD_193230_2019-08-05T18:15:16-04:00_FIVE POINTS",
					time.Time{},
				)
			exec.WillReturnResult(sqlmock.NewResult(0, 0))
		})
		JustBeforeEach(func() {
			callErr = repo.SetArrivalTime(
				marta.Direction("N"),
				marta.Line("GOLD"),
				"193230",
				time.Date(2019, time.August, 5, 18, 15, 16, 0, postgres.EasternTime),
				marta.Station("FIVE POINTS"),
				time.Date(2019, time.August, 5, 20, 15, 16, 0, postgres.EasternTime),
				time.Date(2019, time.August, 5, 22, 15, 16, 0, postgres.EasternTime),
			)
		})
		When("the query fails", func() {
			BeforeEach(func() {
				exec.WillReturnError(errors.New("query failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to set arrival time for dir `N` line `GOLD` train `193230` first event moment `2019-08-05T18:15:16-04:00` and station `FIVE POINTS`: query failed"))
			})
		})
		When("all goes well", func() {
			It("succeeds", func() {
				Expect(callErr).To(BeNil())
			})
		})
	})
})
