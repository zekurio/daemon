package scheduler

type Provider interface {

	// Schedule adds a job to the scheduler. The job will be executed
	// according to the given spec. The spec is a string that is
	Schedule(spec any, job func()) (id any, err error)

	// Unschedule removes the given job by its id from the scheduler.
	Unschedule(id any) error

	// Start runs the scheduler cycle.
	Start()

	// Stop cancels the scheduler cycle.
	Stop()
}
