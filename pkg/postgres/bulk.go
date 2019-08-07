package postgres

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"

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

//LoadDir loads all files in the specified directory
func (a BulkLoaderAgent) LoadDir(dir string) (err error) {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		err = errors.Wrapf(err, "failed to read directory contents for path `%s`", dir)
		return
	}

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
