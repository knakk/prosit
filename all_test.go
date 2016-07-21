package prosit

import (
	"sync"

	"github.com/pkg/errors"
)

type memdb struct {
	mu       sync.RWMutex
	projects map[uint64]Project
	jobs     map[uint64]Job
	runs     map[uint64][]Run
}

func newMemDB() *memdb {
	return &memdb{
		projects: make(map[uint64]Project),
		jobs:     make(map[uint64]Job),
		runs:     make(map[uint64][]Run),
	}
}

func (db *memdb) GetProjects() ([]Project, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	projs := make([]Project, 0, len(db.projects))
	for _, p := range db.projects {
		projs = append(projs, p)
	}
	return projs, nil
}

func (db *memdb) GetProject(id uint64) (Project, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	proj, ok := db.projects[id]
	if !ok {
		return proj, errors.New("project not found")
	}
	return proj, nil
}

func (db *memdb) NewProject(p Project) (Project, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	id := uint64(len(db.projects) + 1)
	p.ID = id
	db.projects[id] = p
	return p, nil
}

func (db *memdb) UpdateProject(p Project) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.projects[p.ID]; !ok {
		return errors.New("project not found")
	}
	db.projects[p.ID] = p
	return nil
}

func (db *memdb) DeleteProject(id uint64) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.projects[id]; !ok {
		return errors.New("project not found")
	}
	delete(db.projects, id)
	return nil
}

func (db *memdb) GetJobs() ([]Job, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	jobs := make([]Job, 0, len(db.jobs))
	for _, job := range db.jobs {
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (db *memdb) GetJob(id uint64) (Job, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if _, ok := db.jobs[id]; !ok {
		return Job{}, errors.New("job not found")
	}
	return db.jobs[id], nil
}

func (db *memdb) NewJob(job Job) (Job, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	id := uint64(len(db.jobs) + 1)
	job.ID = id
	db.jobs[id] = job
	db.runs[id] = make([]Run, 0)
	return job, nil
}

func (db *memdb) UpdateJob(job Job) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.jobs[job.ID]; !ok {
		return errors.New("job not found")
	}
	db.jobs[job.ID] = job
	return nil
}

func (db *memdb) DeleteJob(id uint64) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.jobs[id]; !ok {
		return errors.New("job not found")
	}
	delete(db.jobs, id)
	delete(db.runs, id)
	return nil
}

func (db *memdb) GetRunForJob(runID uint64, jobID uint64) (Run, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if _, ok := db.jobs[jobID]; !ok {
		return Run{}, errors.New("job not found")
	}
	if len(db.runs[jobID]) < int(runID)-1 {
		return Run{}, errors.New("run not found")
	}
	return db.runs[jobID][int(runID)-1], nil
}

func (db *memdb) GetNRunsForJob(jobID uint64, n int) ([]Run, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if _, ok := db.jobs[jobID]; !ok {
		return nil, errors.New("job not found")
	}
	var runs []Run
	c := 0
	for i := len(db.runs[jobID]); i > 0; i-- {
		runs = append(runs, db.runs[jobID][i-1])
		c++
		if c == n {
			break
		}
	}
	return runs, nil
}

func (db *memdb) NewRunForJob(jobID uint64) (Run, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.jobs[jobID]; !ok {
		return Run{}, errors.New("job not found")
	}
	run := Run{
		ID: uint64(len(db.runs[jobID]) + 1),
	}
	db.runs[jobID] = append(db.runs[jobID], run)
	return run, nil
}

func (db *memdb) UpdateRunForJob(jobID uint64, run Run) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.jobs[jobID]; !ok {
		return errors.New("job not found")
	}
	if len(db.runs[jobID]) < int(run.ID)-1 {
		return errors.New("run not found")
	}
	db.runs[jobID][int(run.ID)-1] = run
	return nil
}
