package postgres_test

import (
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
				//
			}
		})
		JustBeforeEach(func() {
			callErr = upserter.AddRecordToDatabase(rec)
		})
		When("the eventTime is malformed", func() {})
		When("the check for the latest matching run fails", func() {})
		When("ensuring the arrival record fails", func() {
			When("the latest run is stale", func() {})
		})
		When("the train has arrived", func() {
			When("setting the arrival time fails", func() {})
			When("all goes well", func() {})
		})
		When("the train has not arrived", func() {
			When("the next arrival time is malformed", func() {})
			When("adding the arrival estimate fails", func() {})
			When("all goes well", func() {})
		})
	})
})
