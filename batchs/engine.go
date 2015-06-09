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
		ctx.move_rename(name, "processing", "error")
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
		ctx.move_rename(name, "processing", "error")
	} else {
		err2 = ctx.move_rename(name, "processing", "success")
	}
	// ignore err since it is only if there was a processing error
	// and a processing error doesn't mean start() errored.
	return err2
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
		// exit loop when only non directories are present
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
				return err
			}
		}
	}
	return nil
}

const (
	pollDuration = 5 * time.Second
)

// watchDir watches the directory dirname and returns when either files are
// added or there is an error.
func (ctx *Context) watchDir(dirname string) error {
	dname := path.Join(ctx.basepath, dirname)
	info, err := os.Stat(dname)
	if err != nil {
		return err
	}

	// since NFS mounts don't support ionotify events, lets poll
	for {
		time.Sleep(pollDuration)
		var newInfo os.FileInfo
		newInfo, err = os.Stat(dname)
		if err != nil {
			break
		}
		if newInfo.ModTime().After(info.ModTime()) {
			break
		}
	}
	return err
}

// Resolve a job name into a path.
// Returns the full path name, and the "directory" the job is in, or an error
// if the name could not be resolved for some reason.
func (ctx *Context) resolve(name string) (fullpath, directory string, err error) {
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

// move a job between status directories
func (ctx *Context) move(name, srcdir, dstdir string) error {
	src := path.Join(ctx.basepath, srcdir, name)
	dst := path.Join(ctx.basepath, dstdir, name)
	return os.Rename(src, dst)
}

// Move a job between status directories, possibly renaming if there
// is a directory inside dstdir having the same name as `name`.
func (ctx *Context) move_rename(name, srcdir, dstdir string) error {
	// If we first check for the existence of a file and then
	// move the directory in two steps, there is a possibility
	// of a race condition, so we just keep trying to move
	// the directory using different source names until one works.
	var count = 1
	src := path.Join(ctx.basepath, srcdir, name)
	dst := path.Join(ctx.basepath, dstdir, name)
	err := os.Rename(src, dst)
	for err != nil {
		dst = path.Join(ctx.basepath, dstdir, fmt.Sprintf("%s-%03d", name, count))
		count++
		err = os.Rename(src, dst)
		if count >= 1000 {
			break
		}
	}
	return err
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

const (
	numRetries = 10
	retryDelay = 10 * time.Second
)

// Run ensures the required directories are present,
// and returns any old processing jobs to the queue.
// It will then watch the queue directory, running jobs when needed.
//
// To account with NFS filesystems, it will retry operations on errors,
// up to a maximum of 10 times. If too many errors happen in a row, this
// function will return.
func (ctx *Context) Run() error {
	const timeout = retryDelay * time.Duration(numRetries+1)
	var errCount int
	var mark = time.Now()
	for {
		err := ctx.run()
		if err == nil {
			// run is returning without error??? okay dokey.
			return nil
		}
		log.Println("Run:", err)
		if time.Now().After(mark) {
			// mark has expired. reset everything.
			errCount = 0
			mark = time.Now().Add(timeout)
			log.Printf("Resetting mark to %v\n", mark)
		}
		errCount++
		log.Println("Error Count:", errCount)
		if errCount >= numRetries {
			return err
		}
		time.Sleep(retryDelay)
	}
}

// run the main event loop forever, returning whenever there is an error
func (ctx *Context) run() error {
	err := ctx.initDirs()
	if err != nil {
		return err
	}
	for {
		// see if there are any jobs in the queue
		if err := ctx.scanAndLoad(); err != nil {
			log.Println("scanAndLoad:", err)
			return err
		}

		// wait for jobs to appear
		if err := ctx.watchDir("queue"); err != nil {
			// an error?
			log.Println("watchDir:", err)
			return err
		}
	}
}

// NewContext creates a new context structure.
// taskpath is used to resolve task names into commands.
func NewContext(basepath, taskpath string) *Context {
	ctx := &Context{basepath: basepath, taskpath: taskpath}
	return ctx
}

func (ctx *Context) initDirs() error {
	// do the directories exist?
	for _, subdir := range subdirs {
		err := os.MkdirAll(path.Join(ctx.basepath, subdir), 0744)
		if err != nil {
			return err
		}
	}

	// return any jobs in processing directory to the queue
	fnames, err := ctx.listJobs("processing")
	if err != nil {
		return err
	}
	for _, finfo := range fnames {
		err = ctx.move(finfo.Name(), "processing", "queue")
		if err != nil {
			return err
		}
	}
	return nil
}
