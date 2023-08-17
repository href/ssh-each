package stream

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStdoutStderr(t *testing.T) {
	i := Result{}

	i.resultType = StdoutResult
	i.output = "Stdout"
	assert.Equal(t, i.Stdout(), "Stdout")
	assert.Equal(t, i.Stderr(), "")

	i.resultType = StderrResult
	i.output = "Stderr"
	assert.Equal(t, i.Stdout(), "")
	assert.Equal(t, i.Stderr(), "Stderr")
}

func TestExitCode(t *testing.T) {
	i := Result{}

	i.resultType = StdoutResult
	assert.Panics(t, func() { i.ExitCode() })

	i.resultType = ExitResult
	assert.Equal(t, 0, i.ExitCode())

	i.exitCode = 1
	assert.Equal(t, 1, i.ExitCode())
}
