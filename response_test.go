package sphinx

import (
	"bytes"
	"testing"
)

// Test get expected errors with a buffer that's too short
func TestGetErrorWithBadBuffer(t *testing.T) {
	// t.Fail()
}

// Do some basic sanity checking on round-tripping of responses
func TestRoundTripInt(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatal(r)
		}
	}()
	const testInteger = 124
	var b bytes.Buffer
	var w = &SafeWriter{
		internalBuf: &b,
		err:         nil,
	}

	w.AddIntToBuffer(testInteger)
	if w.err != nil {
		t.Errorf("Unexpected error writing integer ")
	}

	var r = &ResponseReader{
		Buffer:      &b,
		internalErr: nil,
	}

	if recoveredInteger := r.ReadInt(); recoveredInteger != testInteger {
		t.Errorf(
			"Expected to get integer %v from buffer, got %v\n",
			testInteger, recoveredInteger,
		)
	}

}

// Basic sanity checking on round-tripping of string
func TestRoundTripString(t *testing.T) {
}
