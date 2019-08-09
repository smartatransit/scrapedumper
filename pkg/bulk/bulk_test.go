package bulk_test

import (
	"github.com/bipol/scrapedumper/pkg/postgres"
	"github.com/bipol/scrapedumper/pkg/postgres/postgresfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BulkLoader", func() {
	var (
		upserter *postgresfakes.FakeUpserter

		loader postgres.BulkLoader
	)

	BeforeEach(func() {
		upserter = &postgresfakes.FakeUpserter{}
	})

	JustBeforeEach(func() {
		loader = postgres.NewBulkLoader(upserter)
	})

	Describe("Load", func() {
		var (
		//
		)
	})

	Describe("LoadDir", func() {
	})
})
