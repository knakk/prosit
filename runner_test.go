package prosit

import (
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Verify that a jobs ared executed and the Run metadata stored.
func TestRunJob(t *testing.T) {
	tests := []struct {
		cmd     string
		output  string
		success bool
	}{
		{
			cmd:     "echo 'hi' > delete_me.txt",
			output:  "",
			success: true,
		},
		{
			cmd:     "cat delete_me.txt",
			output:  "hi\n",
			success: true,
		},
		{
			cmd:     "rm delete_me.txt",
			output:  "",
			success: true,
		},
		{
			cmd:     "cat delete_me.txt",
			output:  "cat: delete_me.txt: No such file or directory\nexit status 1",
			success: false,
		},
		{
			cmd:     "echo 'one\ntwo\nthree'; 2>&1 echo 'four';",
			output:  "one\ntwo\nthree\nfour\n",
			success: true,
		},
		{
			cmd:     "echo 'I will fail'; exit 1",
			output:  "I will fail\nexit status 1",
			success: false,
		},
	}

	r := NewRunner(newMemDB())

	for _, test := range tests {
		job, err := r.NewJob(Job{
			Workspace: "/tmp",
			Cmd:       test.cmd,
		})
		if err != nil {
			t.Fatal(err)
		}

		run, err := r.runJob(job.ID, false)
		if err != nil {
			t.Fatal(err)
		}

		runs, err := r.GetNRunsForJob(job.ID, 1)
		if err != nil {
			t.Fatal(err)
		}

		if len(runs) != 1 {
			t.Fatal("run metedata not stored")
		}

		stored := runs[0]
		if stored != run {
			t.Errorf("stored run => %v; expected %v", stored, run)
		}

		if run.End.Before(run.Start) {
			t.Error("run end time is before start time")
		}

		if run.Success != test.success {
			t.Errorf("run result=%v; want %v", run.Success, test.success)
		}

		if run.Output != test.output {
			t.Errorf("output from job %q\ngot %q; want %q", run.Cmd, run.Output, test.output)
		}

		if run.Cmd != test.cmd {
			t.Errorf("stored command was %q; want %q", run.Cmd, test.cmd)
		}
	}
}

// Verify that a Job's output from all previous runs can be retrieved.
func TestRunOutputsForJob(t *testing.T) {
	db := newMemDB()
	r := NewRunner(db)

	job, err := db.NewJob(Job{
		Cmd: "date +%s%N",
	})
	if err != nil {
		t.Fatal(err)
	}

	const numRuns = 10
	for i := 0; i <= numRuns; i++ {
		r.runJob(job.ID, false)
	}

	runs, err := r.GetNRunsForJob(job.ID, numRuns)
	if err != nil {
		t.Fatal(err)
	}

	if len(runs) != numRuns {
		t.Fatalf("only %d runs stored, but job run %d times", len(runs), numRuns)
	}

	// verify that runs where run sequentially in time
	times := make([]int, numRuns)
	for i, run := range runs {
		n, err := strconv.Atoi(strings.TrimSpace(run.Output))
		if err != nil {
			t.Fatal(err)
		}
		times[numRuns-(i+1)] = n
	}

	if !sort.IntsAreSorted(times) {
		t.Fatalf("runs not run sequentially")
	}

}
