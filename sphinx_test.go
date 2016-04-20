package sphinx

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func deserializeRequestBody(b *bytes.Buffer) <-chan string {
	outputChannel := make(chan string)
	go func() {
		var word = make([]byte, 4)
		var hexRepresentation = make([]string, 4)

		for byteLen, err := b.Read(word); err == nil && byteLen == 4; {
			for i := range word {
				hexRepresentation[i] = hex.EncodeToString(word[i : i+1])
			}

			outputChannel <- strings.Join(hexRepresentation, ":")
			byteLen, err = b.Read(word)
		}
		close(outputChannel)
	}()
	return outputChannel
}

// TestFixtureRequests compares each line of each fixture file to a request buffer
// that we generate.  We need to get the file name, generate our own fixture data
// from the query, index, and comment in the file name, and compare line-by-line
// to the data from the fixture file.
func TestFixtureRequests(t *testing.T) {
	const (
		prefix = "fixture_data/generated/"
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

	for _, file := range files {
		t.Logf("Testing fixture data for file %v\n", file.Name())
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

		// Compare each line of the file to each line of the generated hex data
		fixtureFile, err := os.Open(prefix + file.Name())
		if err != nil {
			t.Errorf("Could not open file %v for reading: %v.\n",
				prefix+file.Name(),
				err,
			)
		}
		fixtureLines := bufio.NewScanner(fixtureFile)

		line := 0
		for hexLine := range deserializeRequestBody(buf) {
			line++
			// We're leaking a goroutine but it doesn't matter - quitting test anyway
			if !fixtureLines.Scan() {
				t.Errorf(
					"Error %v on line %v of our request buffer - no more lines available to "+
						"read, but still have lines from the fixture file.",
					fixtureLines.Err(),
					line,
				)
				break
			}
			fixtureText := fixtureLines.Text()
			t.Logf("%v\t%v\n", fixtureText, hexLine)
			if fixtureText != hexLine {
				t.Errorf(
					"Mismatch on line %v: \n%v\ndoes not match\n%v\n", line, fixtureText, hexLine,
				)
			}
		}

		if fixtureLines.Scan() {
			t.Errorf(
				"Still have line %v (and others?) to read from fixture data, but no "+
					"more from the generated request body.\n",
				fixtureLines.Text(),
			)
		}

		fixtureFile.Close()

	}

}

func TestRequestsWithFilters(t *testing.T) {
	// TODO: Need to have separate folder for testing queries with filters
	// t.Fail()
}

func TestRequestsWithWeights(t *testing.T) {
	// TODO: Need to test index and field weights
	// t.Fail()
}
