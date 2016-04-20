package sphinx

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func DeserializeRequestBody(b *bytes.Buffer) <-chan string {
	outputChannel := make(chan string)
	go func() {
		var word = make([]byte, 4)
		var hexRepresentation = make([]string, 4)

		for byteLen, err := b.Read(word); err != nil && byteLen == 4; {
			for i := 0; i < len(word); i++ {
				hexRepresentation[i] = hex.EncodeToString(word)
			}

			outputChannel <- strings.Join(hexRepresentation, ":")
			byteLen, err = b.Read(word)
		}
		close(outputChannel)
	}()
	return outputChannel
}

func TestFixtureRequests(t *testing.T) {
	const (
		prefix = "fixture_data/generated"
		suffix = ".tst"
	)
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

	// For each fixture file, we need to get the file name, generate our own fixture data,
	// and compare line-by-line to the data from the fixture file.
	for _, file := range files {
		fileBaseName := strings.TrimSuffix(file.Name(), "_")
		fileParts := strings.Split(fileBaseName, "_")

		// Normalize for missing comment and replace 'ALL' for index with '*'
		if len(fileParts) == 2 {
			fileParts = append(fileParts, "")
		}

		if fileParts[1] == "ALL" {
			fileParts[1] = "*"
		}

		q := DefaultQuery()
		q.Keywords = fileParts[0]
		q.Index = fileParts[1]
		q.Comment = fileParts[2]

		buf, err := buildInternalQuery(q)
		if err != nil {
			t.Errorf(
				"Could not build request buffer for input query `%v` - got error `%v`\n",
				buf, err,
			)
			continue
		}

		line := 0
		for hexLine := range DeserializeRequestBody(buf) {
			line++
			t.Logf("Line %v of request body: %v",
				line,
				hexLine,
			)
		}
	}

}

func TestRequestsWithFilters(t *testing.T) {
	// Need to have separate folder for testing queries with filters
	// t.Fail()
}

func TestRequestsWithWeights(t *testing.T) {
	// Need to test index and field weights
	// t.Fail()
}
