package stream

import (
	"context"
	"os/exec"
	"sync"
	"time"
)

// CommandResult attaches a command to each result, so we cann tell the origin
// apart when receving results from Mux.
type CommandResult struct {
	Command *exec.Cmd
	Result  Result
}

// Mux multiplexes multiple commands over a set number of workers, and streams
// results back through a single channel, shared by multiple commands.
type Mux struct {
	// ctx carries the context to cancel execution
	ctx context.Context

	// cmds is the source of exec.Cmd instances to execute
	cmds chan *exec.Cmd

	// cmdresults is where (partial) results are received
	cmdresults chan CommandResult

	// mu locks done/left for bookkeeping
	mu *sync.Mutex

	// workers running
	workers uint

	// shut delivers a true once no new cmds should be accepted
	shut chan bool
}

// NewMux starts a new mux with the given amount of workers. Each worker is
// able to process a single command from start to finish.
func NewMux(ctx context.Context, workers uint) *Mux {
	m := Mux{
		ctx:        ctx,
		cmds:       make(chan *exec.Cmd),
		cmdresults: make(chan CommandResult),
		mu:         &sync.Mutex{},
		shut:       make(chan bool),
	}

	go func() {
		select {
		case <-m.ctx.Done():
		case <-m.shut:
		}
		close(m.cmds)
	}()

	for w := uint(0); w < workers; w++ {
		go m.worker()
	}
	return &m
}

// Submit sends command to the command channel. This may block if all workers
// are currently busy. For a non-blocking alternative, see TrySubmit. Submit
// returns true if the command was submitted, false if the context was
// cancelled before that happened.
//
// Note that submitting the same command twice is not supported and results
// in undefined behavior - there is currently no facility that protects you
// from such an error. The same goes for using the command instance in other
// places concurrently.
//
// In other words: If you submit a command the Mux expects to be the sole
// owner of it. Since this is not Rust, the ownership is not enforced however.
func (m *Mux) Submit(cmd *exec.Cmd) bool {
	return ContextSend(m.ctx, m.cmds, cmd)
}

// TrySubmit tries to send a command to the command channel. If the commmand
// is accepted by a worker within 1ms, true is returned.
func (m *Mux) TrySubmit(cmd *exec.Cmd) bool {
	return ContextSendWithTimeout(m.ctx, m.cmds, cmd, 1*time.Millisecond)
}

// Shut stops Mux from accepting more commands. This causes the workers to
// wind down and ensures that the Results channel eventually closes.
func (m *Mux) Shut() {
	m.shut <- true
	close(m.shut)
}

// Results yields the command results as they become available. This channel
// will be open forever, unless Shut is called.
func (m *Mux) Results() <-chan CommandResult {
	return m.cmdresults
}

// worker is the implementation of a single worker
func (m *Mux) worker() {

	// keep track of workers
	m.mu.Lock()
	m.workers++
	m.mu.Unlock()

	// Send results for commands to the channel available to consumers through
	// the Results method.
	for cmd := range m.cmds {
		for result := range StreamCommand(m.ctx, cmd) {
			sent := ContextSend(
				m.ctx,
				m.cmdresults,
				CommandResult{
					Command: cmd,
					Result:  result,
				},
			)

			// If the send was cancelled, we can stop
			if !sent {
				break
			}
		}
	}

	// The last worker shuts the cmdresults so that consumers of those routines
	// know that all is done.
	m.mu.Lock()
	m.workers--
	if m.workers == 0 {
		close(m.cmdresults)
	}
	m.mu.Unlock()
}
