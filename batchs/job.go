package batchs

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"time"
)

// a Job consists of, essentially, a list of tasks.
type Job struct {
	name     string
	log      io.WriteCloser
	path     string
	Owner    string
	Todo     []string
	Finished []Task
}

type Task struct {
	Name   string
	Start  time.Time
	Finish time.Time
	Status string
}

// Set up an envrionment and execute the first task on the Todo list.
// The task is removed from the Todo list and a Task structure is appended
// to the Finished list.
func (jb *Job) executeTask() error {
	tskname := jb.Todo[0]
	t := Task{Name: tskname}
	fmt.Fprintf(jb.log, "\n===== Task %s\n", tskname)

	e := exec.Command("bash", "-c", tskname)
	e.Env = []string{
		fmt.Sprintf("OWNER=%s", jb.Owner),
	}
	e.Stdout = jb.log
	e.Stderr = jb.log
	e.Dir, _ = filepath.Abs(jb.path)

	t.Start = time.Now()
	e.Run()
	t.Finish = time.Now()

	t.Status = "ok"
	fmt.Fprintf(jb.log, "===== End %s\n", tskname)
	jb.Todo = jb.Todo[1:]
	jb.Finished = append(jb.Finished, t)
	return nil
}

// process all the tasks in a Job until either there are no more tasks
// or there is an error.
// (How does it notify the parent context when it is finished?)
func (jb *Job) process() error {
	// take next task and try to run it
	var err error
	for len(jb.Todo) > 0 {
		if err = jb.executeTask(); err != nil {
			break
		}
	}
	return err
}
