// This is a helper program to generate request fixture files with a common naming
// scheme while using the `fixture_data` program to hook into the
// modified C Sphinx library.
package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strings"
)

const TestCaseFile = "fixture_data/test_cases.tsv"

func main() {
	testCaseData, err := os.Open(TestCaseFile)
	if err != nil {
		log.Fatalf(
			"Could not open file %v for reading\n",
			TestCaseFile,
		)
	}
	defer testCaseData.Close()

	scanner := bufio.NewScanner(testCaseData)
	for scanner.Scan() {
		// Generate output file with name based on test case data
		// Body of file will be
		metaData := strings.Split(scanner.Text(), "\t")

		if !(len(metaData) >= 2) {
			log.Fatalf(
				"Expected query, index, and comment in query, got: \n%v\n",
				metaData,
			)
		}

		fileName := strings.Join(metaData, "_")
		fileName = strings.Replace(fileName, "*", "ALL", -1)
		fileName = "fixture_data/generated/" + fileName + ".tst"

		// File already exists, so skip
		if _, err := os.Stat(fileName); err == nil {
			continue
		}

		fixtureFile, err := os.Create(fileName)

		if err != nil {
			log.Fatalf("Could not create fixture file: %v\n", err)
		}

		fixtureCommand := exec.Command("./fixture_data/fixture_data", metaData...)
		// Need to pipe stdout to file
		fixtureCommand.Stdout = fixtureFile

		if err = fixtureCommand.Run(); err != nil {
			log.Fatalf("Error generating fixture data: `%v`\n", err)
		}

		fixtureFile.Close()
	}
}
