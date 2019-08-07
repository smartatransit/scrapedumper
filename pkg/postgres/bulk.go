package postgres

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/pkg/errors"
)

//BulkLoader loads a file into the Postgres database
type BulkLoader interface {
	Load(io.Reader) error
	LoadDir(dir string) error
}

//BulkLoaderAgent implements BulkLoader
type BulkLoaderAgent struct {
	upserter Upserter
}

//NewBulkLoader creates a new BulkLoader
func NewBulkLoader(
	upserter Upserter,
) BulkLoader {
	return BulkLoaderAgent{
		upserter: upserter,
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

//LoadDir loads all files in the specified directory
func (a BulkLoaderAgent) LoadDir(dir string) (err error) {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		err = errors.Wrapf(err, "failed to read directory contents for path `%s`", dir)
		return
	}

	//sort files alphabetically, since their names are RFC3339 timestamps
	sort.Sort(fileInfoList(infos))

	for _, info := range infos {
		if info.IsDir() {
			continue
		}

		var file *os.File
		path := path.Join(dir, info.Name())
		file, err = os.Open(path)
		if err != nil {
			err = errors.Wrapf(err, "failed to open file `%s` for reading", path)
			return
		}

		err = a.Load(file)
		if err != nil {
			err = errors.Wrapf(err, "failed to load contents of file `%s`", path)
			return
		}
	}

	return nil
}

//Load loads a file into the Postgres database
func (a BulkLoaderAgent) Load(r io.Reader) (err error) {
	var records []martaapi.Schedule
	err = json.NewDecoder(r).Decode(&records)
	if err != nil {
		err = errors.Wrap(err, "failed to decode file contents")
		return
	}

	for _, rec := range records {
		err = a.upserter.AddRecordToDatabase(rec)
		if err != nil {
			err = errors.Wrapf(err, "failed to add record `%s` to database", rec.String())
			return
		}
	}

	return nil
}
