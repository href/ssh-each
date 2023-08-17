// Package stream provides a way to execute exec/os.Cmd and return stdout,
// stderr, error, and exit code through a channel that delivers these partial
// results as soon as they become available.
//
// Note that no assumptions about the output buffering are made. You get the
// output of each stdout/stderr write call as an Result in the output channel.
// Usually this means you get output line-by-line (as most stdout/stderr output
// is written that way, especially if an interactive terminal is involved), but
// if we receive parts of a line, we will send that part as a single Result.
package stream

import (
	"context"
	"errors"
	"io"
	"os/exec"
)

// source indicates the output a pipe is attached to.
type source uint8

const (
	fromStdout source = iota + 1
	fromStderr
)

// stream is used to handle output from an os/exec.Cmd and to push it into
// a channel. Each stream is attached to stdout or stderr (not both).
type stream struct {
	source source
	ctx    context.Context
	ch     chan Result
}

// Write implements the io.Write interface so we can attach the stream to
// an os/exec.Cmd.Stdout/Stderr field.
func (s *stream) Write(bytes []byte) (int, error) {
	i := Result{output: string(bytes)}
	if s.source == fromStdout {
		i.resultType = StdoutResult
	} else {
		i.resultType = StderrResult
	}

	if ContextSend(s.ctx, s.ch, i) {
		return len(bytes), nil
	} else {
		return 0, io.ErrClosedPipe
	}
}

// StreamCommand takes a exec.Cmd, starts it, and streams the result through
// the returned channel. Once the command is over, the channel will be
// closed.
func StreamCommand(ctx context.Context, cmd *exec.Cmd) <-chan Result {
	ch := make(chan Result)

	cmd.Stdout = &stream{source: fromStdout, ctx: ctx, ch: ch}
	cmd.Stderr = &stream{source: fromStderr, ctx: ctx, ch: ch}

	// If the command doesn't start in the first place, return immediately
	err := cmd.Start()
	if err != nil {
		go func() {
			defer close(ch)

			ContextSend(ctx, ch, Result{
				err:        err,
				resultType: ErrorResult,
			})
		}()

		return ch
	}

	// Otherwise await the commands end
	go func() {
		defer close(ch)

		err := cmd.Wait()
		var exitErr *exec.ExitError
		switch {
		case err == nil:
			ContextSend(ctx, ch, Result{
				exitCode:   0,
				resultType: ExitResult,
			})
		case errors.As(err, &exitErr):
			ContextSend(ctx, ch, Result{
				exitCode:   exitErr.ExitCode(),
				resultType: ExitResult,
			})
		default:
			ContextSend(ctx, ch, Result{
				err:        err,
				resultType: ErrorResult,
			})
		}
	}()

	return ch
}
