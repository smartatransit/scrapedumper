package dumper_test

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/bipol/scrapedumper/pkg/dumper"
	"github.com/bipol/scrapedumper/pkg/dumper/dumperfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

var _ = Describe("Dump", func() {
	Context("RoundRobinDump", func() {
		var (
			fake1  *dumperfakes.FakeDumper
			fake2  *dumperfakes.FakeDumper
			logger *zap.Logger
			client dumper.Dumper
			r      io.Reader
			err    error
		)
		BeforeEach(func() {
			fake1 = &dumperfakes.FakeDumper{}
			fake2 = &dumperfakes.FakeDumper{}
			logger = zap.NewNop()
			r = strings.NewReader("ahhhhh")
			err = nil
		})
		JustBeforeEach(func() {
			client = dumper.NewRoundRobinDumpClient(logger, fake1, fake2)
			err = client.Dump(context.Background(), r, "some path")
		})
		When("one of the clients error", func() {
			BeforeEach(func() {
				fake1.DumpReturns(errors.New("an error"))
			})
			It("does err", func() {
				Expect(err).To(MatchError("an error"))
			})
		})
		When("it dumps", func() {
			It("does not err", func() {
				Expect(err).To(BeNil())
			})
			It("calls both clients", func() {
				Expect(fake1.DumpCallCount()).To(Equal(1))
				Expect(fake2.DumpCallCount()).To(Equal(1))
			})
		})
	})
	Context("S3DumpClient", func() {
		var (
			uploader *dumperfakes.FakeUploader
			logger   *zap.Logger
			client   dumper.Dumper
			r        io.Reader
			err      error
		)
		BeforeEach(func() {
			uploader = &dumperfakes.FakeUploader{}
			logger = zap.NewNop()
			r = strings.NewReader("ahhhhh")
			err = nil
		})
		JustBeforeEach(func() {
			client = dumper.NewS3DumpClient(uploader, "bucket", logger)
			err = client.Dump(context.Background(), r, "some path")
		})
		When("it dumps", func() {
			It("does not err", func() {
				Expect(err).To(BeNil())
			})
			It("gives the correct upload input", func() {
				inp, _ := uploader.UploadArgsForCall(0)
				Expect(inp).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Bucket": PointTo(Equal("bucket")),
					"Key":    PointTo(Equal("some path")),
				})))
			})
		})
	})
	Context("LocalDumpClient", func() {
		var (
			fs     afero.Fs
			logger *zap.Logger
			client dumper.Dumper
			r      io.Reader
			err    error
		)
		BeforeEach(func() {
			fs = afero.NewMemMapFs()
			logger = zap.NewNop()
			r = strings.NewReader("ahhhhh")
			err = nil
		})
		JustBeforeEach(func() {
			client = dumper.NewLocalDumpClient("path", logger, fs)
			err = client.Dump(context.Background(), r, "somepath")
		})
		When("it dumps", func() {
			It("does not err", func() {
				Expect(err).To(BeNil())
			})
			It("writes to the filesystem", func() {
				_, err := fs.Stat("path/somepath")
				Expect(err).To(BeNil())
			})
		})
	})
})
