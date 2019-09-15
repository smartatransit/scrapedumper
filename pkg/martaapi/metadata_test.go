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

	When("neither orientation scores", func() {
		BeforeEach(func() {
			seq = []Station{FivePointsStation}
		})
		It("returns -, -", func() {
			Expect(line).To(Equal(Line("")))
			Expect(dir).To(Equal(Direction("")))
		})
	})
	When("both orientations score", func() {
		BeforeEach(func() {
			seq = []Station{
				LindberghStation,
				FivePointsStation,
				OmniDomeStation,
			}
		})
		It("returns -, -", func() {
			Expect(line).To(Equal(Line("")))
			Expect(dir).To(Equal(Direction("")))
		})
	})
	When("only E-W registers a score", func() {
		When("the dominant direction is east", func() {
			BeforeEach(func() {
				seq = []Station{
					OmniDomeStation,
					FivePointsStation,
					InmanParkStation,
				}
			})
			It("returns Blue, East", func() {
				Expect(line).To(Equal(Blue))
				Expect(dir).To(Equal(East))
			})
		})
		When("the dominant direction is west", func() {
			BeforeEach(func() {
				seq = []Station{
					InmanParkStation,
					FivePointsStation,
					OmniDomeStation,
				}
			})
			It("returns Blue, West", func() {
				Expect(line).To(Equal(Blue))
				Expect(dir).To(Equal(West))
			})
		})
		When("there's a dead tie", func() {
			BeforeEach(func() {
				seq = []Station{
					InmanParkStation,
					FivePointsStation,
					InmanParkStation,
				}
			})
			It("returns Blue, -", func() {
				Expect(line).To(Equal(Blue))
				Expect(dir).To(Equal(Direction("")))
			})
		})
	})
	When("only N-S registers a score", func() {
		When("neither gold nor red registers a score", func() {
			When("the dominant direction is north", func() {
				BeforeEach(func() {
					seq = []Station{
						FivePointsStation,
						NorthAveStation,
						LindberghStation,
					}
				})
				It("returns -, North", func() {
					Expect(line).To(Equal(Line("")))
					Expect(dir).To(Equal(North))
				})
			})
			When("the dominant direction is south", func() {
				BeforeEach(func() {
					seq = []Station{
						LindberghStation,
						NorthAveStation,
						FivePointsStation,
					}
				})
				It("returns -, South", func() {
					Expect(line).To(Equal(Line("")))
					Expect(dir).To(Equal(South))
				})
			})
			When("there's a dead tie", func() {
				BeforeEach(func() {
					seq = []Station{
						LindberghStation,
						NorthAveStation,
						FivePointsStation,
						LindberghStation,
					}
				})
				It("returns -, -", func() {
					Expect(line).To(Equal(Line("")))
					Expect(dir).To(Equal(Direction("")))
				})
			})
		})
		When("both gold and red register a score", func() {
			BeforeEach(func() {
				seq = []Station{
					FivePointsStation,
					DoravilleStation,
					BuckheadStation,
					FivePointsStation,
				}
			})
			It("returns -, -", func() {
				Expect(line).To(Equal(Line("")))
				Expect(dir).To(Equal(Direction("")))
			})
		})
		When("only gold registers a score", func() {
			BeforeEach(func() {
				seq = []Station{
					FivePointsStation,
					DoravilleStation,
					FivePointsStation,
				}
			})
			It("returns Gold, -", func() {
				Expect(line).To(Equal(Gold))
				Expect(dir).To(Equal(Direction("")))
			})
		})
		When("only red registers a score", func() {
			BeforeEach(func() {
				seq = []Station{
					FivePointsStation,
					BuckheadStation,
					FivePointsStation,
				}
			})
			It("returns Red, -", func() {
				Expect(line).To(Equal(Red))
				Expect(dir).To(Equal(Direction("")))
			})
		})
	})
})
