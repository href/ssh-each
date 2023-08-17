package stream

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMux(t *testing.T) {
	ctx := context.Background()

	m := NewMux(ctx, 2)

	foo := exec.CommandContext(ctx, "echo", "foo")
	bar := exec.CommandContext(ctx, "echo", "bar")

	assert.True(t, m.Submit(foo))
	assert.True(t, m.Submit(bar))

	m.Shut()

	results := make(map[*exec.Cmd][]Result)
	for r := range m.Results() {
		results[r.Command] = append(results[r.Command], r.Result)
	}

	assert.Equal(t, []Result{
		{resultType: StdoutResult, output: "foo\n"},
		{resultType: ExitResult, exitCode: 0},
	}, results[foo])

	assert.Equal(t, []Result{
		{resultType: StdoutResult, output: "bar\n"},
		{resultType: ExitResult, exitCode: 0},
	}, results[bar])
}
