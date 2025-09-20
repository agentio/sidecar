// Package track is used to measure execution times.
package track

import (
	"io"
	"time"
)

func Measure(start time.Time, name string, count int, out io.Writer) {
	if count > 1 {
		elapsed := time.Since(start)
		_, _ = out.Write([]byte(time.Duration(elapsed / time.Duration(count)).String()))
	}
}
