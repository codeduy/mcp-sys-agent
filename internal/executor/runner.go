package executor

// RunResult holds the output of a command execution attempt.
type RunResult struct {
	Output  string
	Err     error
	Timeout bool
}
