package worker_test

import (
	"context"
	"io/ioutil"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/bipol/scrapedumper/pkg/circuitbreaker"
	"github.com/bipol/scrapedumper/pkg/dumper/dumperfakes"
	"github.com/bipol/scrapedumper/pkg/martaapi/martaapifakes"
	"github.com/bipol/scrapedumper/pkg/worker"
	. "github.com/bipol/scrapedumper/pkg/worker"
	"github.com/bipol/scrapedumper/pkg/worker/workerfakes"
)

var _ = Describe("Client", func() {
	Context("WorkList", func() {
		var (
			workList *WorkList
		)
		Context("AddWork", func() {
			BeforeEach(func() {
				workList = worker.NewWorkList()
				for i := 0; i < 5; i++ {
					workList.AddWork(&martaapifakes.FakeScheduleFinder{}, &dumperfakes.FakeDumper{})
				}
			})
			When("adding work", func() {
				It("has the correct length", func() {
					Expect(len(workList.GetWork())).To(Equal(5))
				})
			})
		})
	})
	Context("Poll", func() {
		var (
			workList *workerfakes.FakeWorkGetter
			pollTime time.Duration
			logger   *zap.Logger
			s        worker.ScrapeAndDumpClient
			ctx      context.Context
			opts     []worker.Option
		)
		BeforeEach(func() {
			ctx = context.Background()
			workList = &workerfakes.FakeWorkGetter{}
			logger = zap.NewNop()
			pollTime = 500 * time.Millisecond
			opts = []worker.Option{}
		})
		JustBeforeEach(func() {
			errC := make(chan error, 1)
			s = worker.New(pollTime, logger, workList, opts...)
			s.Poll(ctx, errC)
		})
		When("context is cancelled", func() {
			var cancelFunc context.CancelFunc
			BeforeEach(func() {
				pollTime = 1 * time.Hour
				ctx, cancelFunc = context.WithCancel(ctx)
				cancelFunc()
			})
			It("does not call work", func() {
				Expect(workList.GetWorkCallCount()).To(BeZero())
			})
		})
		When("with a circuit breaker", func() {
			var (
				sc *martaapifakes.FakeScheduleFinder
				d  *dumperfakes.FakeDumper
			)
			BeforeEach(func() {
				sc = &martaapifakes.FakeScheduleFinder{}
				d = &dumperfakes.FakeDumper{}
				sc.FindSchedulesReturns(ioutil.NopCloser(strings.NewReader("")), nil)
				workList.GetWorkReturns([]ScrapeDump{ScrapeDump{Scraper: sc, Dumper: d}})
				cb := circuitbreaker.New(logger, 1*time.Hour, 10)
				opts = append(opts, worker.WithCircuitBreaker(cb))
			})
			It("scrapes and dumps", func() {
				Eventually(func() int { return sc.FindSchedulesCallCount() }).Should(BeNumerically(">=", 1))
				Eventually(func() int { return d.DumpCallCount() }).Should(BeNumerically(">=", 1))
			})

		})
		When("given work", func() {
			var (
				sc *martaapifakes.FakeScheduleFinder
				d  *dumperfakes.FakeDumper
			)
			BeforeEach(func() {
				sc = &martaapifakes.FakeScheduleFinder{}
				d = &dumperfakes.FakeDumper{}
				sc.FindSchedulesReturns(ioutil.NopCloser(strings.NewReader("")), nil)
				workList.GetWorkReturns([]ScrapeDump{ScrapeDump{Scraper: sc, Dumper: d}})
			})
			It("scrapes and dumps", func() {
				Eventually(func() int { return sc.FindSchedulesCallCount() }).Should(BeNumerically(">=", 1))
				Eventually(func() int { return d.DumpCallCount() }).Should(BeNumerically(">=", 1))
			})
		})
	})
})
