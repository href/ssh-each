package term

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// HasStdin returns true if an stdin file is attached
func HasStdin() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// StdinReader returns a reader that reads from stdin
func StdinReader() io.Reader {
	return bufio.NewReader(os.Stdin)
}

// CommaSeparatedReader returns a reader where every item found in a
// comma separated string is returned as a new line.
func CommaSeparatedReader(items string) io.Reader {
	return strings.NewReader(strings.ReplaceAll(items, ",", "\n"))
}

// CombinedReader returns a reader that will first return one line for each
// item in the given comma-separated string, followed by the lines read from
// STDIN.
//
// If no items are given and no stdin is attached, nil is returned.
func CombinedReader(items string) io.Reader {
	var readers [2]io.Reader

	if HasStdin() {
		readers[0] = StdinReader()
	}

	if items != "" {
		readers[1] = CommaSeparatedReader(items)
	}

	if readers[0] != nil && readers[1] != nil {
		return io.MultiReader(readers[0], readers[1])
	}

	if readers[0] != nil {
		return readers[0]
	}

	if readers[1] != nil {
		return readers[1]
	}

	return nil
}
