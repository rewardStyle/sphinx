package sphinx

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFixtureRequests(t *testing.T) {
	const prefix = "fixture_data/generated"
	// Get fixture data from generated fixture directory and
	var (
		files []os.FileInfo
		err   error
	)

	if files, err = ioutil.ReadDir(prefix); err != nil {
		t.Fatalf(
			"Could not read fixture files: `%v`\n",
			err,
		)
	}

	// For each fixture file, we need to
	for _, file := range files {
		_ = file.Name()
	}
}

func TestRequestsWithFilters(t *testing.T) {
	// t.Fail()
}
