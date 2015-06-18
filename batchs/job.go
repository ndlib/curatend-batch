package batchs

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Job collects all the logical consists of a single processing job.
// It is, essentially, a list of tasks.
type Job struct {
	name     string         // the id of this job
	log      *log.Logger    // logging interface to this job's LOG
	logfile  io.WriteCloser // and the underlying file object (if applicable)
	path     string         // location of this job on disk. passed to tasks
	Todo     []string       // FIFO list of task names to execute
	Finished []Task         // list of finished tasks, from earliest to latest
	taskpath string         // location of task command files
	state    State          // the current state of this job
}

// Task is a singular work item inside a Job.
type Task struct {
	Name   string
	Start  time.Time
	Finish time.Time
	Status string
}

// State encodes the processing status of a Job.
type State int

const (
	// StateUnknown is the "zero state"
	StateUnknown State = iota
	// StateQueue means job is awaiting processing
	StateQueue
	// StateProcessing means job is being processed
	StateProcessing
	// StateSuccess means job is finished and was successful
	StateSuccess
	// StateError means job is done being processed and there was an error
	StateError
)

// Set up an envrionment and execute the given task name.
func (jb *Job) executeTask(tskname string) Task {
	t := Task{Name: tskname,
		Status: "Error", // assume an error
	}
	t.Start = time.Now()

	controlf, err := ioutil.TempFile("", "curate-task-")
	if err != nil {
		jb.log.Println("Error creating control file:", err)
		t.Finish = time.Now()
		return t
	}
	controlfname := controlf.Name()
	controlf.Close()
	defer os.Remove(controlfname)

	tcommand := jb.resolve(tskname)
	if tcommand == "" {
		jb.log.Printf("Could not resolve task '%s'", tskname)
		t.Finish = time.Now()
		return t
	}
	jb.log.Printf("exec '%s'", tcommand)
	e := exec.Command(tcommand)
	e.Env = []string{
		fmt.Sprintf("JOBPATH=%s", jb.path),
		fmt.Sprintf("JOBNAME=%s", jb.name),
		fmt.Sprintf("JOBCONTROL=%s", controlfname),
	}
	e.Stdout = jb.logfile
	e.Stderr = jb.logfile
	e.Dir, err = filepath.Abs(jb.path)

	if err == nil {
		err = e.Run()
	}
	t.Finish = time.Now()

	if err == nil {
		err = jb.readControl(controlfname)
	}

	if err != nil {
		jb.log.Println(err)
	} else {
		t.Status = "ok"
	}
	return t
}

// Given a task name, return a command to execute, or
// empty string if no task could be resolved
func (jb *Job) resolve(taskname string) string {
	tcommand, err := filepath.Abs(filepath.Join(jb.taskpath, taskname))
	if err != nil {
		return ""
	}
	_, err = os.Stat(tcommand)
	if err != nil {
		return ""
	}
	return tcommand
}

// reads the given control file, and mutates jb appropriately.
func (jb *Job) readControl(fname string) error {
	body, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	lines := strings.Split(string(body), "\n")
	for _, ln := range lines {
		switch {
		case ln == "":
			// skip blank lines
		case strings.HasPrefix(ln, "addtask:"):
			jb.Todo = append(jb.Todo, ln[8:])
		default:
			return fmt.Errorf("malformed control message %s", ln)
		}
	}
	return nil
}

var (
	// ErrTask means a task had a non-zero exit code
	ErrTask = errors.New("error running a task")
)

// Process all the tasks in a Job until either there are no more tasks
// or there is an error. Tasks are removed from the Todo list as they
// are successfully completed.
func (jb *Job) process() error {
	// take next task and try to run it
	for len(jb.Todo) > 0 {
		tskname := jb.Todo[0]
		jb.log.Printf("===== Task %s", tskname)
		t := jb.executeTask(tskname)
		jb.log.Printf("===== End %s", tskname)
		jb.Finished = append(jb.Finished, t)

		jb.log.Printf("===== Status: %s", t.Status)
		if t.Status != "ok" {
			return ErrTask
		}

		// only remove task if successful
		jb.Todo = jb.Todo[1:]
	}
	return nil
}
