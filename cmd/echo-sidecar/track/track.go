// Package track is used to measure execution times.
package track

import (
	"fmt"
	"io"
	"time"
)

func Measure(start time.Time, name string, count int, out io.Writer) {
	if count > 1 {
		elapsed := time.Since(start)
		out.Write([]byte(fmt.Sprintf("%s", elapsed/time.Duration(count))))
	}
}
