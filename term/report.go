package term

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/href/ssh-each/stream"
)

// ReportMode refers to the various output possiblities that Report supports.
type ReportMode uint8

const (
	// HostReport prints the hostname before each line of output.
	HostReport = iota + 1

	// PlainReport shows only the output from the SSH commands.
	PlainReport

	// CheckReport shows the hostname and a ✓ or x depending on the exit code.
	CheckReport

	// ExitReport shows the hostname and the exit code.
	ExitReport

	// SilentReport suppresses all output
	SilentReport
)

const (
	MinReport = HostReport
	MaxReport = SilentReport
)

// ReportModeFromString returns the ReportMode for the given string. Uses the
// comma ok idiom to indicate if that worked.
func ReportModeFromString(mode string) (ReportMode, bool) {
	switch mode {
	case "host":
		return HostReport, true
	case "plain":
		return PlainReport, true
	case "check":
		return CheckReport, true
	case "exit":
		return ExitReport, true
	case "silent":
		return SilentReport, true
	default:
		return 0, false
	}
}

// Report accepts stream.CommandResult instances, keeps track of their status
// and offers various output modes.
type Report struct {
	exitCodes    []int
	successCodes map[int]bool
	mode         ReportMode
	mu           *sync.Mutex
	registry     map[*exec.Cmd]string
}

// NewReport creates a new report.
func NewReport(mode ReportMode) Report {
	if !(MinReport <= mode && mode <= MaxReport) {
		panic(fmt.Sprintf("unsupported mode: %d", mode))
	}
	return Report{
		mode:         mode,
		exitCodes:    make([]int, 0),
		mu:           &sync.Mutex{},
		registry:     make(map[*exec.Cmd]string),
		successCodes: map[int]bool{0: true},
	}
}

// Associate links a string to a exec.Cmd pointer. Usually the name will be
// the name of the server the command is run on.
func (r *Report) Associate(name string, cmd *exec.Cmd) {
	r.mu.Lock()
	r.registry[cmd] = name
	r.mu.Unlock()
}

// Success indicates if the run set of commands were a success.
func (r *Report) Success() bool {
	// If no command ran, this is not a success
	if len(r.exitCodes) == 0 {
		return false
	}

	// Otherwise, we have success if all commands have a successful exit code
	for i := 0; i < len(r.exitCodes); i++ {
		if !r.successCodes[r.exitCodes[i]] {
			return false
		}
	}
	return true
}

// On is given a command result, which it tracks and outputs according to the
// report mode set.
func (r *Report) On(cmdresult stream.CommandResult) {
	r.mu.Lock()
	server := r.registry[cmdresult.Command]
	r.mu.Unlock()

	result := cmdresult.Result

	r.mu.Lock()
	defer r.mu.Unlock()
	switch result.Type() {
	case stream.StdoutResult:
		r.printOutput(os.Stdout, server, result.Stdout())
	case stream.StderrResult:
		r.printOutput(os.Stderr, server, result.Stderr())
	case stream.ErrorResult:
		fmt.Fprint(os.Stderr, server, "error:", result.Err())
	case stream.ExitResult:
		r.exitCodes = append(r.exitCodes, result.ExitCode())
		r.printResult(server, result.ExitCode())
	}
}

// printOutput prints the given result if it's a stdout/stderr output.
func (r *Report) printOutput(file *os.File, server string, output string) {
	if output == "" {
		return
	}

	switch r.mode {
	case SilentReport:
		return
	case CheckReport:
		return
	case ExitReport:
		return
	case PlainReport:
		fmt.Fprint(file, output)
	case HostReport:
		prefix := func() {
			fmt.Fprint(file, server, ": ")
		}
		printPrefix := true
		for _, char := range output {
			if r.mode == HostReport && printPrefix {
				prefix()
				printPrefix = false
			}

			fmt.Fprint(file, string(char))
			if char == '\n' {
				printPrefix = true
			}
		}
	default:
		panic(fmt.Sprintf("unsupported mode: %d", r.mode))
	}
}

// printResult prints the given result if it is an ExitResult
func (r *Report) printResult(server string, exitCode int) {
	switch r.mode {
	case SilentReport:
		return
	case PlainReport:
		return
	case HostReport:
		return
	case CheckReport:
		var mark string

		if r.successCodes[exitCode] {
			mark = "✓"
		} else {
			mark = "x"
		}

		fmt.Fprint(os.Stdout, server, ": ", mark, "\n")
	case ExitReport:
		fmt.Fprint(os.Stdout, server, ": ", exitCode, "\n")
	default:
		panic(fmt.Sprintf("unsupported mode: %d", r.mode))
	}
}
