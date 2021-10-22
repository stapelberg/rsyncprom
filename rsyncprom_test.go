package rsyncprom

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	got, err := Parse(strings.NewReader(`

sent 1,192,097 bytes  received 1,039 bytes  795,424.00 bytes/sec
total size is 1,188,046  speedup is 1.00
`))
	if err != nil {
		t.Fatal(err)
	}
	want := &Stats{
		Found:        true,
		TotalWritten: 1192097,
		TotalRead:    1039,
		BytesPerSec:  795424,
		TotalSize:    1188046,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parse(): unexpected output: diff (-want +got):\n%s", diff)
	}
}

func ExampleParse() {
	stats, err := Parse(strings.NewReader(`

sent 1,192,097 bytes  received 1,039 bytes  795,424.00 bytes/sec
total size is 1,188,046  speedup is 1.00
`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("corpus is %d bytes big!\n", stats.TotalSize)
	// Output: corpus is 1188046 bytes big!
}
