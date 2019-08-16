package bulk_test

import (
	"context"
	"errors"
	"os"

	"github.com/bipol/scrapedumper/pkg/bulk"
	"github.com/bipol/scrapedumper/pkg/bulk/bulkfakes"
	"github.com/bipol/scrapedumper/pkg/dumper/dumperfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func fInfo(name string, dir bool) *bulkfakes.FakeFileInfo {
	fi := &bulkfakes.FakeFileInfo{}
	fi.NameReturns(name)
	fi.IsDirReturns(dir)
	return fi
}

var _ = Describe("DirectoryDumperAgent", func() {
	var (
		fs     *bulkfakes.FakeFileSystem
		dumper *dumperfakes.FakeDumper

		agent bulk.DirectoryDumperAgent
	)

	BeforeEach(func() {
		fs = &bulkfakes.FakeFileSystem{}
		dumper = &dumperfakes.FakeDumper{}
	})

	JustBeforeEach(func() {
		agent = bulk.NewDirectoryDumper(fs, dumper)
	})

	Describe("DumpDirectory", func() {
		var (
			dirFile *bulkfakes.FakeFile

			callErr error
		)
		BeforeEach(func() {
			dirFile = &bulkfakes.FakeFile{}

			dirFile.ReaddirReturns([]os.FileInfo{
				fInfo("1", false),
				fInfo("4", false),
				fInfo("3", false),
				fInfo("2", true),
			}, nil)

			fs.OpenReturnsOnCall(0, dirFile, nil)
			fs.OpenReturnsOnCall(1, &bulkfakes.FakeFile{}, nil)
			fs.OpenReturnsOnCall(2, &bulkfakes.FakeFile{}, nil)
			fs.OpenReturnsOnCall(3, &bulkfakes.FakeFile{}, nil)
		})

		JustBeforeEach(func() {
			callErr = agent.DumpDirectory(context.Background(), "/path/to/dir")
		})

		When("the directory can't be opened", func() {
			BeforeEach(func() {
				fs.OpenReturnsOnCall(0, nil, errors.New("open failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to open directory contents for reading at path `/path/to/dir`: open failed"))
			})
		})
		When("the directory can't be read", func() {
			BeforeEach(func() {
				dirFile.ReaddirReturns(nil, errors.New("readdir failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to read directory contents for path `/path/to/dir`: readdir failed"))
			})
		})
		When("a file can't be opened", func() {
			BeforeEach(func() {
				fs.OpenReturnsOnCall(1, nil, errors.New("open failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to open file `/path/to/dir/1` for reading: open failed"))
			})
		})
		When("a file can't be dumped", func() {
			BeforeEach(func() {
				dumper.DumpReturns(errors.New("dump failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError("failed to dump contents of file `/path/to/dir/1`: dump failed"))
			})
		})
		When("all goes well", func() {
			It("succeeds", func() {
				Expect(callErr).To(BeNil())

				Expect(fs.OpenCallCount()).To(Equal(4))
				Expect(fs.OpenArgsForCall(0)).To(Equal("/path/to/dir"))
				Expect(fs.OpenArgsForCall(1)).To(Equal("/path/to/dir/1"))
				Expect(fs.OpenArgsForCall(2)).To(Equal("/path/to/dir/3"))
				Expect(fs.OpenArgsForCall(3)).To(Equal("/path/to/dir/4"))
			})
		})
	})
})
