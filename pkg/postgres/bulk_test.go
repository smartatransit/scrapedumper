package postgres_test

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
		//TODO create a fake reader and feed it
		//different hard-coded JSON strings
	})

	Describe("LoadDir", func() {
		//TODO gonna need to refactor it to use
	})
})
