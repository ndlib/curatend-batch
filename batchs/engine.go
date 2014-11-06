package batchs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

// Context tracks the path of where we store the queues.
// It is the only structure which knows about how the queues are persisted
// to disk.
// It should be promoted to an interface, probably.
type Context struct {
	basepath string // the path to the base directory holding our state directories
	taskpath string // the path to the directory mapping task names to commands
}

// load a job from disk. will look for the job in the directory dir,
// or pass "" to search all directories.
func (ctx *Context) load(dir, name string) (*Job, error) {
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
	result := &Job{name: name}
	// open the log first, so if there is an error there is a place to put it
	result.logfile, err = os.OpenFile(path.Join(jpath, "LOG"),
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0664)
	if err != nil {
		log.Println("load:", err)
		return nil, err
	}
	result.log = log.New(result.logfile, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	result.path = jpath
	result.taskpath = ctx.taskpath
	f, err := os.Open(path.Join(jpath, "JOB"))
	if err != nil {
		if os.IsNotExist(err) {
			// it is okay if there is no JOB file
			return result, nil
		}
		// not okay if there is a JOB file and we cannot read it
		result.log.Println("Error opening JOB file:", err)
		result.logfile.Close()
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(result)
	if err != nil {
		result.log.Println("Error reading JOB file:", err)
		result.logfile.Close()
		return nil, err
	}

	return result, nil
}

func (ctx *Context) save(dir string, jb *Job) error {
	var err error
	jpath := path.Join(ctx.basepath, dir, jb.name)
	f, err := os.Create(path.Join(jpath, "JOB"))
	if err != nil {
		jb.log.Println("Error saving JOB file:", err)
		return err
	}
	defer f.Close()

	data, err := json.MarshalIndent(jb, "", "    ")
	if err == nil {
		_, err = f.Write(data)
	}

	if err != nil && jb.log != nil {
		jb.log.Println("Error saving JOB file:", err)
		if jb.logfile != nil {
			jb.logfile.Close()
			jb.logfile = nil
		}
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

	// If a job has no tasks assigned and has completed no tasks
	// seed it with the default task `start`
	if len(jb.Todo)+len(jb.Finished) == 0 {
		jb.Todo = append(jb.Todo, "start")
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
// everything.
func (ctx *Context) scanAndLoad() error {
	// count the number of dentries which are not directories
	var nFiles int
	for {
		dentries, err := ctx.listJobs("queue")
		if err != nil {
			return err
		}
		// exit for loop when only non directories are present
		if len(dentries) <= nFiles {
			break
		}
		nFiles = 0
		for _, finfo := range dentries {
			if !finfo.IsDir() {
				nFiles++
				continue
			}
			log.Printf("Processing job %s", finfo.Name())
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
// taskpath is used to resolve task names into commands.
func NewContext(basepath, taskpath string) *Context {
	ctx := &Context{basepath: basepath, taskpath: taskpath}
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
