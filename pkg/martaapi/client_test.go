package martaapi_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/bipol/scrapedumper/pkg/martaapi/martaapifakes"
)

var _ = Describe("Client", func() {
	var (
		doer      *martaapifakes.FakeDoer
		apiKey    string
		client    Client
		resp      *http.Response
		retErr    error
		err       error
		schedules []Schedule
	)
	BeforeEach(func() {
		doer = new(martaapifakes.FakeDoer)
		apiKey = "apikey"
		err = nil
	})
	JustBeforeEach(func() {
		doer.DoReturns(resp, retErr)
		client = Client{
			Doer:   doer,
			ApiKey: apiKey,
		}
	})
	Context("FindSchedules", func() {
		JustBeforeEach(func() {
			schedules, err = client.FindSchedules()
		})
		When("marta returns an invalid body", func() {
			BeforeEach(func() {
				resp = &http.Response{
					Body: ioutil.NopCloser(bytes.NewBufferString("k")),
				}
			})
			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})
		When("marta returns a non 200", func() {
			BeforeEach(func() {
				retErr = errors.New("some api err")
			})
			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})
		When("marta returns a valid schedule", func() {
			When("the schedule has no records", func() {
				BeforeEach(func() {
					resp = &http.Response{
						Body: ioutil.NopCloser(bytes.NewBufferString("[]")),
					}
				})
				It("returns an empty response", func() {
					Expect(schedules).To(BeEmpty())
				})
				It("does not error", func() {
					Expect(err).To(BeNil())
				})
			})
			When("the schedule has several records", func() {
				BeforeEach(func() {
					resp = &http.Response{
						Body: ioutil.NopCloser(bytes.NewBufferString(ValidScheduleJSON)),
					}
				})
				It("does not error", func() {
					Expect(err).To(BeNil())
				})
				It("returns the correct schedules", func() {
					Expect(schedules).To(Equal(ValidScheduleExpectation))
				})
			})
		})
	})

})
