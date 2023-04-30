package codeexec

import "time"

type Payload struct {
	Language string
	Code     string
	// TODO add other fields maybe
}

type Response struct {
	StdOut   string
	StdErr   string
	ExecTime time.Duration
	MemUsed  string
	CPUTime  string
}

// ExecutorWrapper is our interface for code execution
type ExecutorWrapper interface {
	HasSupport(lang string) bool

	NewExecutor(guildID string) (Executor, error)
}

type Executor interface {
	Execute(Payload) (Response, error)
}
