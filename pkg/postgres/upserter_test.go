package postgres_test

import (
	"errors"
	"time"

	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/bipol/scrapedumper/pkg/postgres"
	"github.com/bipol/scrapedumper/pkg/postgres/postgresfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dump", func() {
	var (
		repo *postgresfakes.FakeRepository

		upserter postgres.Upserter
	)

	BeforeEach(func() {
		repo = &postgresfakes.FakeRepository{}
	})

	JustBeforeEach(func() {
		upserter = postgres.NewUpserter(repo, 10*time.Minute)
	})

	Context("AddRecordToDatabase", func() {
		var (
			rec     martaapi.Schedule
			callErr error
		)
		BeforeEach(func() {
			rec = martaapi.Schedule{
				Direction:   "N",
				Line:        "GOLD",
				Destination: "DORAVILLE STATION",
				TrainID:     "324898",
				Station:     "FIVE POINTS STATION",
				EventTime:   "6/18/2019 9:41:02 PM",
				NextArrival: "9:45:02 PM",
			}

			repo.GetLatestRunStartMomentForReturns(
				time.Date(2019, time.June, 18, 21, 42, 2, 0, postgres.EasternTime),
				time.Date(2019, time.June, 18, 21, 43, 2, 0, postgres.EasternTime),
				nil,
			)
		})
		JustBeforeEach(func() {
			callErr = upserter.AddRecordToDatabase(rec)
		})
		When("the eventTime is malformed", func() {
			BeforeEach(func() {
				rec.EventTime = "asdf"
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to parse record event time `asdf`: parsing time \"asdf\": month out of range"))
			})
		})
		When("the check for the latest matching run fails", func() {
			BeforeEach(func() {
				repo.GetLatestRunStartMomentForReturns(time.Time{}, time.Time{}, errors.New("query failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to get latest run start moment for record `N:GOLD:DORAVILLE STATION:324898:6/18/2019 9:41:02 PM:false`: query failed"))
			})
		})
		When("ensuring the arrival record fails", func() {
			BeforeEach(func() {
				repo.EnsureArrivalRecordReturns(errors.New("query failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to ensure pre-existing arrival record for `N:GOLD:DORAVILLE STATION:324898:6/18/2019 9:41:02 PM:false`: query failed"))
				_, _, _, runStartMoment, _ := repo.EnsureArrivalRecordArgsForCall(0)
				Expect(runStartMoment).To(Equal(time.Date(2019, time.June, 18, 21, 42, 2, 0, postgres.EasternTime)))
			})

			When("the latest run is stale", func() {
				BeforeEach(func() {
					repo.GetLatestRunStartMomentForReturns(
						time.Time{},
						time.Date(2019, time.June, 18, 21, 43, 2, 0, postgres.EasternTime),
						nil,
					)
				})
				It("fails", func() {
					Expect(callErr).To(MatchError("failed to ensure pre-existing arrival record for `N:GOLD:DORAVILLE STATION:324898:6/18/2019 9:41:02 PM:false`: query failed"))

					_, _, _, runStartMoment, _ := repo.EnsureArrivalRecordArgsForCall(0)
					Expect(runStartMoment).To(Equal(time.Date(2019, time.June, 18, 21, 41, 2, 0, postgres.EasternTime)))
				})
			})
		})
		When("the train has arrived", func() {
			BeforeEach(func() {
				rec.WaitingTime = "Arriving"
			})
			When("setting the arrival time fails", func() {
				BeforeEach(func() {
					repo.SetArrivalTimeReturns(errors.New("query failed"))
				})
				It("fails", func() {
					Expect(callErr).To(MatchError("failed to set arrival time from record `N:GOLD:DORAVILLE STATION:324898:6/18/2019 9:41:02 PM:true`: query failed"))
				})
			})
		})
		When("the train has not arrived", func() {
			When("the next arrival time is malformed", func() {
				BeforeEach(func() {
					rec.NextArrival = "asdf"
				})
				It("fails", func() {
					Expect(callErr).To(MatchError("failed to parse record estimated arrival time `asdf`: parsing time \"asdf\" as \"3:04:05 PM\": cannot parse \"asdf\" as \"3\""))
				})
			})
			When("adding the arrival estimate fails", func() {
				BeforeEach(func() {
					repo.AddArrivalEstimateReturns(errors.New("query failed"))
				})
				It("fails", func() {
					Expect(callErr).To(MatchError("failed to add arrival estimate from record `N:GOLD:DORAVILLE STATION:324898:6/18/2019 9:41:02 PM:false`: query failed"))
				})
			})
			When("all goes well", func() {
				It("succeeds", func() {
					Expect(callErr).To(BeNil())

					_, _, _, _, _, _, estimate := repo.AddArrivalEstimateArgsForCall(0)
					Expect(estimate).To(Equal(time.Date(2019, time.June, 18, 21, 45, 2, 0, postgres.EasternTime)))
				})
			})
		})
	})
})
