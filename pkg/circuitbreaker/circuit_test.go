package circuitbreaker_test

import (
	"time"

	"github.com/pkg/errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	. "github.com/bipol/scrapedumper/pkg/circuitbreaker"
)

var _ = Describe("Circuit", func() {
	Context("CircuitBreaker", func() {
		var (
			cb       *CircuitBreaker
			waitTime time.Duration
			window   int
		)
		Context("Run", func() {
			var (
				err error
			)
			BeforeEach(func() {
				window = 5
				err = nil
				cb = New(zap.NewNop(), waitTime, window)
			})

			When("we fail window size excessively", func() {
				JustBeforeEach(func() {
					for i := 0; i < window*2; i++ {
						err = cb.Run(func() error { return errors.New("") })
					}
				})
				It("returns an open circuit error", func() {
					Expect(errors.Cause(err)).To(MatchError(ErrSystemFailure))
				})
			})
			When("we fail window size", func() {
				JustBeforeEach(func() {
					for i := 0; i < window-1; i++ {
						err = cb.Run(func() error { return errors.New("") })
						Expect(err).To(BeNil())
					}
					err = cb.Run(func() error { return errors.New("") })
				})
				It("returns an open circuit error", func() {
					Expect(err).To(MatchError(ErrOpenCircuit))
				})
				When("we then recover", func() {
					JustBeforeEach(func() {
						for i := 0; i < window; i++ {
							err = cb.Run(func() error { return nil })
							Expect(err).To(BeNil())
						}
					})
					It("returns a half open circuit", func() {
						Expect(err).To(BeNil())
					})
				})
			})
		})
	})
	Context("BooleanRollingWindow", func() {
		var (
			bw   *BooleanRollingWindow
			size int
		)
		Context("All", func() {
			var (
				b bool
				s bool
			)
			BeforeEach(func() {
				size = 2
				bw = NewBooleanWindow(size)
			})
			JustBeforeEach(func() {
				b = bw.All(s)
			})
			When("we search for false", func() {
				When("with a value", func() {
					BeforeEach(func() {
						bw.Add(true)
					})
					It("returns true", func() {
						Expect(b).To(BeFalse())
					})
				})

				When("we have no values", func() {
					BeforeEach(func() {
						s = false
					})
					It("returns false", func() {
						Expect(b).To(BeTrue())
					})
				})
			})
		})
		Context("Add", func() {
			When("we exceed the size", func() {
				BeforeEach(func() {
					size = 2
					bw = NewBooleanWindow(size)
				})
				JustBeforeEach(func() {
					bw.Add(true)
					bw.Add(true)
					bw.Add(false)
				})
				It("wraps around", func() {
					Expect(bw.Vals).To(ConsistOf(true, false))
				})
			})
			When("we add values", func() {
				JustBeforeEach(func() {
					bw.Add(true)
				})
				BeforeEach(func() {
					size = 5
					bw = NewBooleanWindow(size)
				})
				It("adds the value", func() {
					Expect(bw.Vals).To(ConsistOf(false, false, false, false, true))
				})
			})
		})
	})

})
