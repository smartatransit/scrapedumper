package bulk

import (
	"context"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"github.com/bipol/scrapedumper/pkg/dumper"
)

//DirectoryDumper loads a file into the Postgres database
//go:generate counterfeiter . DirectoryDumper
type DirectoryDumper interface {
	DumpDirectory(ctx context.Context, dir string) error
}

//DirectoryDumperAgent implements DirectoryDumper
type DirectoryDumperAgent struct {
	fs     afero.Fs
	dumper dumper.Dumper
}

//NewDirectoryDumper creates a new DirectoryDumper
func NewDirectoryDumper(
	fs afero.Fs,
	dumper dumper.Dumper,
) DirectoryDumper {
	return DirectoryDumperAgent{
		fs:     fs,
		dumper: dumper,
	}
}

type fileInfoList []os.FileInfo

// Len is the number of elements in the collection.
func (fil fileInfoList) Len() int {
	return len(fil)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (fil fileInfoList) Less(i, j int) bool {
	return strings.Compare(fil[i].Name(), fil[j].Name()) <= 0
}

// Swap swaps the elements with indexes i and j.
func (fil fileInfoList) Swap(i, j int) {
	t := fil[j]
	fil[j] = fil[i]
	fil[i] = t
}

//DumpDirectory loads all files in the specified directory
func (a DirectoryDumperAgent) DumpDirectory(ctx context.Context, dir string) (err error) {

	f, err := a.fs.Open(dir)
	if err != nil {
		err = errors.Wrapf(err, "failed to open directory contents for reading at path `%s`", dir)
		return
	}
	defer f.Close()

	list, err := f.Readdir(-1)
	if err != nil {
		err = errors.Wrapf(err, "failed to read directory contents for path `%s`", dir)
		return
	}

	//sort files alphabetically, since their names are RFC3339 timestamps
	sort.Sort(fileInfoList(list))

	for _, finfo := range list {
		if finfo.IsDir() {
			continue
		}

		var file *os.File
		path := path.Join(dir, finfo.Name())
		file, err = os.Open(path)
		if err != nil {
			err = errors.Wrapf(err, "failed to open file `%s` for reading", path)
			return
		}

		err = a.dumper.Dump(ctx, file, file.Name())
		if err != nil {
			err = errors.Wrapf(err, "failed to dump contents of file `%s`", path)
			return
		}
	}

	return nil
}
