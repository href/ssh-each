package stream

// ResultType desribes the type of a single result coming through the pipe
type ResultType uint8

const (
	// StdoutType is set on results whose text is from Stdout
	StdoutResult = iota + 1

	// StderrType is set on results whose text is from Stderr
	StderrResult

	// ErrorResult is set on results whose error property is set
	ErrorResult

	// ExitResult is set on results that have an exit code
	ExitResult
)

// Result denotes a single data object in the result pipe
type Result struct {
	// resultType is the ResultType of this result.
	resultType ResultType

	// stdout/stderr output, if resultType is StdoutResult/StderrResult
	output string

	// err is set to a non-nil value, if resultType is ErrorResult.
	err error

	// exitCode is set to the exit code if resultType is ExitResult
	exitCode int
}

// Type returns the ResultType of the result
func (i *Result) Type() ResultType {
	return i.resultType
}

// Stdout returns the stdout output if this is an StdoutResult, otherwise
// an empty string is returned.
func (i *Result) Stdout() string {
	if i.resultType == StdoutResult {
		return i.output
	}

	return ""
}

// Stderr returns the stderr output if this is an StderrResult, otherwise
// an empty string is returned.
func (i *Result) Stderr() string {
	if i.resultType == StderrResult {
		return i.output
	}

	return ""
}

// ExitCode returns the exit code, if this is a ExitResult, otherwise
// the function will panic.
func (i *Result) ExitCode() int {
	if i.resultType == ExitResult {
		return i.exitCode
	}

	panic("tried to call ExitCode() on an result of another type")
}

// Err returns the error, if this is an ErrorResult, otherwise
// nil is returned.
func (i *Result) Err() error {
	return i.err
}
