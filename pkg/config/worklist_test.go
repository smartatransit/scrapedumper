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
		cfg.BusDumper = &config.DumpConfig{
			Kind:         config.S3DumperKind,
			S3BucketName: "my-bucket",
		}
		cfg.TrainDumper = &config.DumpConfig{
			Kind:         config.S3DumperKind,
			S3BucketName: "my-bucket",
		}
	})

	JustBeforeEach(func() {
		result, callErr = config.BuildWorkList(nil, cfg, martaapi.Client{}, martaapi.Client{})
	})

	It("builds a worklist", func() {
		Expect(callErr).To(BeNil())
		Expect(result).To(BeAssignableToTypeOf(worker.WorkList{}))
	})

	When("the bus dumper can't be built", func() {
		BeforeEach(func() {
			cfg.BusDumper.S3BucketName = ""
		})
		It("fails", func() {
			Expect(callErr).To(MatchError(ContainSubstring("failed to build bus dumper: dumper kind S3 requested but no s3 bucket name provided")))
		})
	})

	When("the train dumper can't be built", func() {
		BeforeEach(func() {
			cfg.TrainDumper.S3BucketName = ""
		})
		It("fails", func() {
			Expect(callErr).To(MatchError(ContainSubstring("failed to build train dumper: dumper kind S3 requested but no s3 bucket name provided")))
		})
	})
})
