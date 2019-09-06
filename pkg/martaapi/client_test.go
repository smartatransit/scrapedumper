package martaapi_test

import (
	"bytes"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/bipol/scrapedumper/pkg/martaapi"
	. "github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/bipol/scrapedumper/pkg/martaapi/martaapifakes"
)

var _ = Describe("Client", func() {
	var (
		doer   *martaapifakes.FakeDoer
		apiKey string
		client Client
		resp   *http.Response
		retErr error
		err    error
	)
	BeforeEach(func() {
		resp = &http.Response{
			Body: ioutil.NopCloser(bytes.NewBufferString("[]")),
		}
		doer = new(martaapifakes.FakeDoer)
		apiKey = "apikey"
		retErr = nil
		err = nil
	})
	JustBeforeEach(func() {
		doer.DoReturns(resp, retErr)
		logger, _ := zap.NewProduction()
		defer func() {
			_ = logger.Sync()
		}()
		client = New(
			doer,
			apiKey,
			logger,
			"test",
			"prefix",
		)
	})
	Context("New", func() {
		When("called", func() {
			It("does not return an err", func() {
				Expect(err).To(BeNil())
				Expect(client).ToNot(BeNil())
			})
		})
	})
})

var _ = Describe("Schedule", func() {
	Describe("HarArrived", func() {
		It("works", func() {
			Expect(martaapi.Schedule{WaitingTime: "ArrIVIng"}.HasArrived()).To(BeTrue())
			Expect(martaapi.Schedule{WaitingTime: "arrIVED"}.HasArrived()).To(BeTrue())
			Expect(martaapi.Schedule{WaitingTime: "boarDING"}.HasArrived()).To(BeTrue())
			Expect(martaapi.Schedule{WaitingTime: "WHAT"}.HasArrived()).To(BeFalse())
		})
	})
	Describe("String", func() {
		It("works", func() {
			Expect(martaapi.Schedule{Direction: "N"}.String()).NotTo(BeEmpty())
		})
	})
})
