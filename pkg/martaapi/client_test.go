package martaapi_test

import (
	"bytes"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

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
		err = nil
		retErr = nil
	})
	JustBeforeEach(func() {
		doer.DoReturns(resp, retErr)
		logger, _ := zap.NewProduction()
		defer logger.Sync()
		client = New(
			doer,
			apiKey,
			logger,
			"test",
		)
	})
})
