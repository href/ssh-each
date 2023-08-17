package stream

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestStreamCommand(t *testing.T) {
	run := func(cmd *exec.Cmd) []Result {
		results := []Result{}
		for r := range StreamCommand(context.Background(), cmd) {
			results = append(results, r)
		}
		return results
	}

	assert.Equal(t, []Result{
		{
			resultType: StdoutResult,
			output:     "foo\n",
		},
		{
			resultType: ExitResult,
			exitCode:   0,
		},
	}, run(exec.Command("echo", "foo")))

	assert.Equal(t, []Result{
		{
			resultType: StderrResult,
			output:     "bar\n",
		},
		{
			resultType: ExitResult,
			exitCode:   0,
		},
	}, run(exec.Command("sh", "-c", "echo bar >&2")))

	assert.Equal(t, []Result{
		{
			resultType: ExitResult,
			exitCode:   0,
		},
	}, run(exec.Command("true")))

	assert.Equal(t, []Result{
		{
			resultType: ExitResult,
			exitCode:   1,
		},
	}, run(exec.Command("false")))

	assert.Equal(t, []Result{
		{
			resultType: ErrorResult,
			err: &fs.PathError{
				Op:   "fork/exec",
				Path: "/bin/unknown-command",
				Err:  syscall.Errno(2),
			},
		},
	}, run(exec.Command("/bin/unknown-command")))
}

func TestStreamCommandAbort(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := StreamCommand(ctx, exec.CommandContext(ctx, "sleep", "5"))
	cancel()

	start := time.Now()
	result := <-ch
	assert.WithinDuration(t, start, time.Now(), 1*time.Second)
	fmt.Print(result)
}
