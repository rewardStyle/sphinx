package sphinx

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
)

func deserializeRequestBody(b *bytes.Buffer) <-chan string {
	outputChannel := make(chan string)
	go func() {
		var word = make([]byte, 4)

		outputChannel <- strconv.Itoa(b.Len())

		for byteLength, err := b.Read(word); err == nil; byteLength, err = b.Read(word) {
			var hexRepresentation = make([]string, 0, byteLength)
			for i := 0; i < byteLength; i++ {
				hexRepresentation = append(hexRepresentation, hex.EncodeToString(word[i:i+1]))
			}

			output := strings.Join(hexRepresentation, ":")
			// Full 4-byte representation with separators is 11 characters long
			if len(output) != 11 {
				output = output + ":"
			}
			outputChannel <- output
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
		fileBaseName := strings.TrimSuffix(file.Name(), suffix)
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

		if !fixtureLines.Scan() {
			t.Fatalf("Can't read first line (buffer size) from the fixture data.")
		}

		generatedRequestBody := deserializeRequestBody(buf)
		header := <-generatedRequestBody
		fixtureHeader := fixtureLines.Text()

		if header != fixtureLines.Text() {
			t.Errorf(
				"Buffer length mismatch: fixture data gives %v bytes, generated is %v bytes\n",
				fixtureHeader,
				header,
			)
		}

		t.Logf("Buffer length: %v\n", header)
		line := 0
		for hexLine := range generatedRequestBody {
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
			t.Logf("%-11v\t%-11v\n", fixtureText, hexLine)
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

// Tests that can establish client, connect to localhost server, and run basic query
// without error.  Only run if doing long version of tests.
func TestBasicClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Not running simple integration test - don't know if Sphinx configured.")
	} else {
		s := SphinxClient{
			config: DefaultConfig,
		}

		err := s.Init(nil)
		if err != nil {
			t.Errorf("Unexpected error establishing connection : %v\n", err)
		}
		q := DefaultQuery()
		q.Keywords = "test"
		response, err := s.Query(q)
		if err != nil {
			t.Errorf("Unexpected error doing basic query: %v\n", err)
		}
		t.Logf("Got response data: %v\n", response)
		s.Close()
	}
}
