package prosit

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// The Runner coordinates and runs Project jobs.
type Runner struct {
	Storer

	mu           sync.Mutex      // protects the following variables:
	runningJobs  map[uint64]bool // currently running jobs
	runningProjs map[uint64]bool // currently running project pipelines
	scheduled    []task          // scheduled jobs/projects
}

type task struct {
	job  uint64
	proj uint64
}

// NewRunner returns a new Runner, using the given Storer.
func NewRunner(s Storer) *Runner {
	return &Runner{
		Storer:       s,
		runningJobs:  make(map[uint64]bool),
		runningProjs: make(map[uint64]bool),
	}
}

// ScheduleJob schedules a Job. If it is not allready running,
// it will be stared immediately, otherwise queued. An error
// will be returned if the job cannot be retrieved.
func (r *Runner) ScheduleJob(id uint64) error {
	return nil
}

// ScheduleProject schedules a Project. If it is not allready running,
// it will be started immediately, otherwise queued. An error
// will be returned if the Project cannot be retrieved.
func (r *Runner) ScheduleProject(id uint64) error {
	return nil
}

func (r *Runner) runJob(id uint64, inProjectPipeline bool) (Run, error) {
	// TODO why return Run struct?
	defer r.doneJob(id, inProjectPipeline)
	var run Run

	job, err := r.GetJob(id)
	if err != nil {
		return run, errors.Wrapf(err, "cannot get job %d", id)
	}

	run, err = r.NewRunForJob(id)
	if err != nil {
		return run, errors.Wrap(err, "cannot assign run ID")
	}

	if err = createAndChdir(job.Workspace); err != nil {
		return run, errors.Wrap(err, "failed to create/change workspace")
	}

	cmd := exec.Command("/bin/sh", "-c", job.Cmd)
	run.Cmd = job.Cmd

	outF, err := ioutil.TempFile("", "prosit-")
	if err != nil {
		return run, errors.Wrap(err, "failed to create run output file")
	}

	defer os.Remove(outF.Name())
	cmd.Stdout = outF
	cmd.Stderr = outF

	err = cmd.Run()
	if err != nil {
		outF.WriteString(err.Error())
	} else {
		run.Success = true
	}
	run.End = time.Now()
	outF.Close()
	output, err := ioutil.ReadFile(outF.Name())
	if err != nil {
		return run, errors.Wrapf(err, "failed to read temporary output file %s for run %d for job %d", outF.Name(), run.ID, job.ID)
	}
	run.Output = string(output)

	if err := r.UpdateRunForJob(job.ID, run); err != nil {
		return run, errors.Wrapf(err, "failed to update run %d for job %d", run.ID, job.ID)
	}

	return run, nil
}

func (r *Runner) runPipe(id uint64) error {
	proj, err := r.GetProject(id)
	if err != nil {
		return errors.Wrapf(err, "cannot get project %d", id)
	}
	defer r.donePipeline(id)

	for _, jobID := range proj.Pipeline {
		r.mu.Lock()
		// TODO check if job is running first
		r.runningJobs[jobID] = true
		r.mu.Unlock()
		run, err := r.runJob(jobID, true)
		if err != nil {
			return err
		}

		if !run.Success {
			break
		}
	}
	return nil
}

func (r *Runner) doneJob(id uint64, inProjectPipeline bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Mark the job as not running
	delete(r.runningJobs, id)

	if !inProjectPipeline {
		// If there are enqueued jobs/pipelines  - run it
		if len(r.scheduled) > 0 {
			task := r.scheduled[0]
			r.scheduled = r.scheduled[1:]
			if task.job != 0 {
				go func() {
					_, err := r.runJob(task.job, false)
					if err != nil {
						log.Println(err)
					}
				}()
			} else {
				go r.runPipe(task.proj)
			}
		}
	}
}

func (r *Runner) donePipeline(id uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Mark the Project as not running
	delete(r.runningProjs, id)

	// If there are enqueued jobs/pipelines  - run it
	if len(r.scheduled) > 0 {
		task := r.scheduled[0]
		r.scheduled = r.scheduled[1:]
		if task.job != 0 {
			go r.runJob(task.job, false)
		} else {
			go r.runPipe(task.proj)
		}
	}
}

func createAndChdir(path string) error {
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Workspace directory doesn't exist, try to create it
		if err := os.Mkdir(path, 0666); err != nil {
			return fmt.Errorf("failed to create workspace directory: %v", err)
		}
	}

	// Change working directory to job workspace
	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("failed to change to working directory: %v", err)
	}

	return nil
}
