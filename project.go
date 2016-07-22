package prosit

import "time"

// A Project represents a software project that has a number of jobs associated with it.
type Project struct {
	ID   uint64
	Name string

	// The pipeline is a list of jobs that will be run sequentially, stopping if any
	// of the jobs ends with exit status other than 0.
	Pipeline []uint64

	// OneOffJobs is for jobs that is not part of the Pipeline, but can be run individually.
	OneOffJobs []uint64
}

// A Job is a job that can be executed.
type Job struct {
	ID   uint64
	Name string

	// The command to be run. It will be run by /bin/sh.
	Cmd string

	// The workspace where the command will be run. If empty, it will
	// be run in the current working directory of the process.
	Workspace string
}

// Run represents a run of a Job.
type Run struct {
	ID    uint64
	Start time.Time
	End   time.Time
	Cmd   string

	// The combined output to standard out and stadard err.
	Output string

	// A Run is considered successfull if it is not canceled,
	// and the exit code is 0.
	Success bool

	Canceled bool
}
