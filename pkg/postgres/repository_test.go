package postgres_test

import (
	"database/sql"
	"errors"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

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
		repo = postgres.NewRepository(db)
	})

	Describe("EnsureTables", func() {
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
	"most_recent_event_moment" text,
	"direction" text,
	"line" text,
	"train_id" text,
	"run_first_event_moment" text,
	"station" text,
	"arrival_time" text,
	"arrival_estimates" text,
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
SELECT run_first_event_moment, most_recent_event_moment
FROM "arrivals"
WHERE run_group_identifier = \$1
ORDER BY run_first_event_moment DESC, most_recent_event_moment DESC, "arrivals"."identifier" ASC
LIMIT 1`).
				WithArgs("N_GOLD_193230")

			rows = sqlmock.NewRows([]string{"run_first_event_moment", "most_recent_event_moment"})
			query.WillReturnRows(rows)
		})
		JustBeforeEach(func() {
			runFirstEventMoment, mostRecentEventTime, callErr = repo.GetLatestRunStartMomentFor(
				martaapi.Direction("N"),
				martaapi.Line("GOLD"),
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
					easternDate(2019, time.August, 5, 18, 15, 16, 0),
					easternDate(2019, time.August, 5, 18, 34, 16, 0),
				)
			})
			It("succeeds", func() {
				Expect(runFirstEventMoment).To(Equal(easternDate(2019, time.August, 5, 18, 15, 16, 0)))
				Expect(mostRecentEventTime).To(Equal(easternDate(2019, time.August, 5, 18, 34, 16, 0)))
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
INSERT INTO "arrivals"
\("identifier", "run_identifier", "run_group_identifier", "most_recent_event_moment", "direction", "line", "train_id", "run_first_event_moment", "station", "arrival_time", "arrival_estimates"\)
VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, \$10, \$11\)
ON CONFLICT DO NOTHING`).
				WithArgs(
					"N_GOLD_193230_2019-08-05T18:15:16-04:00_FIVE POINTS",
					"N_GOLD_193230_2019-08-05T18:15:16-04:00",
					"N_GOLD_193230",
					easternDate(2019, time.August, 5, 18, 15, 16, 0),
					"N",
					"GOLD",
					"193230",
					easternDate(2019, time.August, 5, 18, 15, 16, 0),
					"FIVE POINTS",
					postgres.EasternTime(time.Time{}),
					postgres.ArrivalEstimates(map[string]string{}),
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
		When("all goes well", func() {
			It("succeeds", func() {
				Expect(callErr).To(BeNil())
			})
		})
	})

	Describe("AddArrivalEstimate", func() {
		var (
			callErr error

			query *sqlmock.ExpectedQuery
			rows  *sqlmock.Rows

			exec      *sqlmock.ExpectedExec
			eventTime postgres.EasternTime

			expectedJSONString string
		)
		BeforeEach(func() {
			query = smock.ExpectQuery(`
SELECT arrival_estimates
FROM "arrivals"
WHERE "arrivals"\."identifier" = \$1
LIMIT 1`).
				WithArgs("N_GOLD_193230_2019-08-05T18:15:16-04:00_FIVE POINTS")

			rows = sqlmock.NewRows([]string{"arrival_estimates"})
			rows.AddRow(`{"2019-08-05T19:15:16-04:00":"2019-08-05T22:15:16-04:00"}`)
			query.WillReturnRows(rows)

			exec = smock.ExpectExec(`
UPDATE "arrivals"
SET \("arrival_estimates", "most_recent_event_moment"\)
  = \(\$1, \$2\)
WHERE "arrivals"\."identifier" = \$3`)
			exec.WillReturnResult(sqlmock.NewResult(0, 0))

			eventTime = easternDate(2019, time.August, 5, 20, 15, 16, 0)
			expectedJSONString = `{"2019-08-05T19:15:16-04:00":"2019-08-05T22:15:16-04:00","2019-08-05T20:15:16-04:00":"2019-08-05T22:15:16-04:00"}`
		})
		JustBeforeEach(func() {
			exec.WithArgs(
				expectedJSONString,
				easternDate(2019, time.August, 5, 20, 15, 16, 0),
				"N_GOLD_193230_2019-08-05T18:15:16-04:00_FIVE POINTS",
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
		When("the query fails", func() {
			BeforeEach(func() {
				query.WillReturnError(errors.New("query failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to get arrival for `N` line `GOLD` and train `193230`: query failed"))
			})
		})
		When("the update fails", func() {
			BeforeEach(func() {
				exec.WillReturnError(errors.New("exec failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to add arrival estimate for dir `N` line `GOLD` train `193230` first event moment `2019-08-05T18:15:16-04:00` and station `FIVE POINTS`: exec failed"))
			})
		})
		When("all goes well", func() {
			When("there's no collision", func() {
				BeforeEach(func() {
					expectedJSONString = `{"2019-08-05T19:15:16-04:00":"2019-08-05T22:15:16-04:00","2019-08-05T20:15:16-04:00":"2019-08-05T22:15:16-04:00"}`
				})
				It("succeeds", func() {
					Expect(callErr).To(BeNil())
				})
			})

			When("the event time already has an estimate recorded", func() {
				BeforeEach(func() {
					eventTime = easternDate(2019, time.August, 5, 19, 15, 16, 0)
				})
				It("fails", func() {
					Expect(callErr).To(BeNil())
				})
			})
		})
	})

	Describe("SetArrivalTime", func() {
		var (
			callErr error

			exec *sqlmock.ExpectedExec
		)
		BeforeEach(func() {
			exec = smock.ExpectExec(`
UPDATE "arrivals"
SET \("arrival_time", "most_recent_event_moment"\)
  = \(\$1, \$2\)
WHERE "arrivals"."identifier" = \$3
  AND "arrival_time" = \$4`).
				WithArgs(
					easternDate(2019, time.August, 5, 22, 15, 16, 0),
					easternDate(2019, time.August, 5, 20, 15, 16, 0),
					"N_GOLD_193230_2019-08-05T18:15:16-04:00_FIVE POINTS",
					postgres.EasternTime(time.Time{}),
				)
			exec.WillReturnResult(sqlmock.NewResult(0, 0))
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
