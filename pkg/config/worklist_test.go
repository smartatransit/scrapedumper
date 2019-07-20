package config_test

import (
	"github.com/bipol/scrapedumper/pkg/config"
	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/bipol/scrapedumper/pkg/worker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildWorkList", func() {
	var (
		cfg config.WorkConfig

		result  worker.WorkList
		callErr error
	)

	BeforeEach(func() {
		cfg.BusDumper = config.DumpConfig{
			Kind: config.RoundRobinKind,
		}
		cfg.TrainDumper = config.DumpConfig{
			Kind: config.RoundRobinKind,
		}
	})

	JustBeforeEach(func() {
		result, callErr = config.BuildWorkList(nil, cfg, martaapi.Client{}, martaapi.Client{})
	})

	It("builds a worklist", func() {
		Expect(callErr).To(BeNil())
		Expect(result).To(BeAssignableToTypeOf(worker.WorkList{}))
	})

	When("bus dumper can't be built", func() {
		BeforeEach(func() {
			cfg.BusDumper.Kind = config.S3DumperKind
		})
		It("fails", func() {
			Expect(callErr).To(MatchError(ContainSubstring("dumper kind S3 requested but no s3 bucket name provided")))
		})
	})

	When("train dumper can't be built", func() {
		BeforeEach(func() {
			cfg.TrainDumper.Kind = config.S3DumperKind
		})
		It("fails", func() {
			Expect(callErr).To(MatchError(ContainSubstring("dumper kind S3 requested but no s3 bucket name provided")))
		})
	})

	When("train dumper can't be built", func() {
		BeforeEach(func() {
			cfg.TrainDumper.Kind = config.S3DumperKind
		})
		It("fails", func() {
			Expect(callErr).To(MatchError(ContainSubstring("dumper kind S3 requested but no s3 bucket name provided")))
		})
	})
})
