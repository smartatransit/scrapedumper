package martaapi

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClassifySequenceList", func() {
	var (
		seq  []Station
		line Line
		dir  Direction
	)

	JustBeforeEach(func() {
		line, dir = ClassifySequenceList(seq, Line(""), Direction(""))
	})

	When("the sequence contains BankheadStation", func() {
		When("BankheadStation is first", func() {
			BeforeEach(func() {
				seq = []Station{
					BankheadStation,
					FivePointsStation,
					EdgewoodCandlerParkStation,
					CollegeParkStation,
					InmanParkStation,
				}
			})
			It("returns Green, West", func() {
				Expect(line).To(Equal(Green))
				Expect(dir).To(Equal(West))
			})
		})
		When("BankheadStation is not first", func() {
			BeforeEach(func() {
				seq = []Station{
					FivePointsStation,
					EdgewoodCandlerParkStation,
					CollegeParkStation,
					InmanParkStation,
					BankheadStation,
				}
			})
			It("returns Green, East", func() {
				Expect(line).To(Equal(Green))
				Expect(dir).To(Equal(East))
			})
		})
	})

	//TODO
})
