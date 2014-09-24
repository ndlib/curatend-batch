package batchs

import (
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

// a Job consists of, essentially, a list of tasks.
type Job struct {
	name     string         // the id of this job
	log      *log.Logger    // logging interface to this job's LOG
	logfile  io.WriteCloser // and the underlying file object (if applicable)
	path     string         // location of this job on disk. passed to tasks
	Owner    string         // the job's owner. passed to tasks
	Todo     []string       // FIFO list of task names to execute
	Finished []Task         // list of finished tasks, from earliest to latest
	taskpath string         // location of task command files
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
	jb.log.Printf("===== Task %s", tskname)
	t.Start = time.Now()

	controlf, err := ioutil.TempFile("", "curate-task-")
	if err != nil {
		jb.log.Println("Error creating control file:", err)
		t.Status = "Error"
		t.Finish = time.Now()
	} else {
		controlfname := controlf.Name()
		controlf.Close()

		tcommand := jb.resolve(tskname)
		if tcommand == "" {
			jb.log.Printf("Could not resolve task '%s'", tskname)
			t.Status = "Error"
			t.Finish = time.Now()
		} else {
			jb.log.Printf("exec '%s'", tcommand)
			e := exec.Command(tcommand)
			e.Env = []string{
				fmt.Sprintf("OWNER=%s", jb.Owner),
				fmt.Sprintf("JOBPATH=%s", jb.path),
				fmt.Sprintf("JOBNAME=%s", jb.name),
				fmt.Sprintf("JOBCONTROL=%s", controlfname),
			}
			e.Stdout = jb.logfile
			e.Stderr = jb.logfile
			e.Dir, err = filepath.Abs(jb.path)

			err = e.Run()
			t.Finish = time.Now()

			if err != nil {
				jb.log.Println(err)
				t.Status = "Error"
			} else {
				t.Status = "ok"
				jb.readControl(controlfname)
			}
		}
		os.Remove(controlfname)
	}

	jb.log.Printf("===== End %s\n", tskname)
	jb.Todo = jb.Todo[1:]
	jb.Finished = append(jb.Finished, t)

	if t.Status == "ok" {
		return nil
	} else {
		return fmt.Errorf("Error")
	}
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
		if !strings.HasPrefix(ln, "addtask:") {
			continue
		}
		jb.Todo = append(jb.Todo, ln[8:])
	}
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