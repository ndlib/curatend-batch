package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"
)

type status int

const (
	statusUnknown    = iota
	statusNew        // job is in queue directory
	statusProcessing // in processing directory
	statusError      // in error directory
	statusDone       // in success directory
)

// a job consists of, essentially, a list of tasks.
type job struct {
	name     string
	status   status
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
func (jb *job) executeTask() error {
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

// process all the tasks in a job until either there are no more tasks
// or there is an error. Expects the task to be in the stateProcessing state.
// (How does it notify the parent context when it is finished?)
func (jb *job) process() error {
	// take next task and try to run it
	if jb.status != statusProcessing {
		panic(fmt.Errorf("Tried to process a job in state %d", jb.status))
	}
	var err error
	for len(jb.Todo) > 0 {
		if err = jb.executeTask(); err != nil {
			break
		}
	}
	return err
}

// Context tracks the path of where we store the queues.
// It is the only structure which knows about how the queues are persisted
// to disk.
// It should be promoted to an interface, probably.
type Context struct {
	basepath string // the path to the base directory holding our state directories
}

// load a job from disk. will look for the job in the directory dir,
// or pass "" to search all directories.
func (ctx *Context) load(dir, name string) (*job, error) {
	var jpath string
	var err error
	if dir == "" {
		jpath, dir, err = ctx.resolve(name)
	} else {
		jpath = path.Join(ctx.basepath, dir, name)
	}
	if err != nil {
		return nil, err
	}
	result := &job{name: name}
	// open the log first, so if there is an error there is a place to put it
	result.log, err = os.OpenFile(path.Join(jpath, "LOG"),
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0664)
	if err != nil {
		log.Println("load:", err)
		return nil, err
	}
	f, err := os.Open(path.Join(jpath, "JOB"))
	if err != nil {
		fmt.Fprintf(result.log, "Error opening JOB file: %s\n", err)
		result.log.Close()
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(result)
	if err != nil {
		fmt.Fprintf(result.log, "Error reading JOB file: %s\n", err.Error())
		result.log.Close()
		return nil, err
	}

	switch dir {
	case "queue":
		result.status = statusNew
	case "processing":
		result.status = statusProcessing
	case "success":
		result.status = statusDone
	case "error":
		result.status = statusError
	}
	result.path = jpath
	return result, nil
}

func (ctx *Context) save(dir string, jb *job) error {
	var err error
	jpath := path.Join(ctx.basepath, dir, jb.name)
	f, err := os.Create(path.Join(jpath, "JOB"))
	if err != nil {
		fmt.Fprintf(jb.log, "Error saving JOB file: %s", err.Error())
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	err = enc.Encode(jb)

	if err != nil && jb.log != nil {
		fmt.Fprintf(jb.log, "Error saving JOB file: %s", err.Error())
		jb.log.Close()
	}
	return err
}

// Start processing the job `name`.
// Expects the job file to be in the queue directory.
// Moves the job to processing, and starts a goroutine to
// process its tasks.
func (ctx *Context) start(name string) error {
	err := ctx.move(name, "queue", "processing")
	if err != nil {
		return err
	}
	jb, err := ctx.load("processing", name)
	if err != nil {
		// try to move it---ignore any errors
		ctx.move(name, "processing", "error")
		return err
	}

	err = jb.process()

	err2 := ctx.save("processing", jb)
	if err != nil || err2 != nil {
		ctx.move(name, "processing", "error")
	} else {
		ctx.move(name, "processing", "success")
	}
	return nil
}

// Scan the queue directory and load jobs one by one until we have processed
// everything. The jobs are loaded in alphabetical order...this may starve jobs
// with names beginning later in the alphabet.
func (ctx *Context) scanAndLoad() error {
	for {
		dentries, err := ctx.listJobs("queue")
		if err != nil {
			return err
		}
		if len(dentries) == 0 {
			break
		}
		for _, finfo := range dentries {
			log.Printf("Found %s", finfo.Name())
			if !finfo.IsDir() {
				log.Printf("Not a directory. Skipping")
				continue
			}
			err = ctx.start(finfo.Name())
			if err != nil {
				log.Println("scanAndLoad:", err)
			}
		}
	}
	return nil
}

const (
	pollDuration = 5 * time.Second
)

// WatchDir returns when new files are added to the directory `dirname` (or an error).
func (ctx *Context) watchDir(dirname string) error {
	dname := path.Join(ctx.basepath, dirname)
	info, err := os.Stat(dname)
	if err != nil {
		return err
	}

	// since NFS mounts don't support ionotify events, lets poll
	tick := time.NewTicker(pollDuration)
loop:
	for {
		select {
		case <-tick.C:
			newInfo, err := os.Stat(dname)
			if err != nil {
				log.Println(err)
			}
			if newInfo.ModTime().After(info.ModTime()) {
				break loop
			}
		}
	}
	tick.Stop()
	return nil
}

// resolve a job name into a path, or error.
// returns the full path name, and the "directory" the job is in
func (ctx *Context) resolve(name string) (string, string, error) {
	for _, subdir := range subdirs {
		s := path.Join(ctx.basepath, subdir, name)
		_, err := os.Stat(s)
		if err == nil {
			// found the file
			return s, subdir, nil
		}
	}
	return "", "", fmt.Errorf("Could not find job: %s", name)
}

// move a job between status directory
func (ctx *Context) move(name, srcdir, dstdir string) error {
	src := path.Join(ctx.basepath, srcdir, name)
	dst := path.Join(ctx.basepath, dstdir, name)
	return os.Rename(src, dst)
}

// list jobs in a given status directory
func (ctx *Context) listJobs(dir string) ([]os.FileInfo, error) {
	sourcedir := path.Join(ctx.basepath, dir)
	return ioutil.ReadDir(sourcedir)
}

var (
	subdirs = []string{
		"queue",
		"processing",
		"success",
		"error",
	}
)

// Run starts a goroutine to watch the queue directory, and goroutines to handle
// the Job processing. The argument is the maximum number of simultaneous jobs.
func (ctx *Context) Run() {
	for {
		// see if there are any jobs in the queue
		if err := ctx.scanAndLoad(); err != nil {
			log.Println("Run:", err)
			break
		}

		// wait for jobs to appear
		if err := ctx.watchDir("queue"); err != nil {
			// an error?
			log.Println("watchDir:", err)
			break
		}
	}
}

// NewContext creates a new context structure, ensures the required directories
// are present, and returns any old processing jobs to the queue.
func NewContext(basepath string) *Context {
	ctx := &Context{basepath: basepath}
	ctx.initDirs()
	return ctx
}

func (ctx *Context) initDirs() {
	// do the directories exist?
	for _, subdir := range subdirs {
		os.MkdirAll(path.Join(ctx.basepath, subdir), 0744)
	}

	// return any jobs in processing directory to the queue
	fnames, err := ctx.listJobs("processing")
	if err != nil {
		log.Println(err)
		return
	}
	for _, finfo := range fnames {
		err = ctx.move(finfo.Name(), "processing", "queue")
		if err != nil {
			log.Println(err)
		}
	}
}
