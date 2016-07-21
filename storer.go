package prosit

// A Storer can persist and retrieve persisted prosit domain types.
type Storer interface {
	// Project methods
	GetProjects() ([]Project, error)
	GetProject(uint64) (Project, error)
	NewProject(Project) (Project, error)
	UpdateProject(Project) error
	DeleteProject(uint64) error

	// Job methods
	GetJobs() ([]Job, error)
	GetJob(uint64) (Job, error)
	NewJob(Job) (Job, error)
	UpdateJob(Job) error
	DeleteJob(uint64) error

	// Job runs methods
	GetRunForJob(uint64, uint64) (Run, error)
	GetNRunsForJob(uint64, int) ([]Run, error)
	NewRunForJob(uint64) (Run, error)
	UpdateRunForJob(uint64, Run) error
}
