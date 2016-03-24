package batchs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

// A JobSource abstracts the way the main processor gets jobs.
type JobSource interface {
	// Perform any initializations
	Init() error
	// Return the next available job. Block until one is available.
	// May return a job AND an error, in which case the job should be finalized.
	NextJob() (*Job, error)
	// Save and finalize the given job
	FinishJob(job *Job) error
}

type fileQueue struct {
	// the path to the base directory holding our state directories
	basepath string
}

// Return the next job available on the file system
func (fq *fileQueue) NextJob() (*Job, error) {
	for {
		job, err := fq.scanAndLoad()
		if job != nil || err != nil {
			return job, err
		}
		// wait for jobs to appear
		if err = fq.watchDir("queue"); err != nil {
			return nil, err
		}
	}
}

// Save the given job
func (fq *fileQueue) FinishJob(jb *Job) error {
	var err2 error
	var sourceDir = fq.findJobDir(jb.name)
	err := fq.save(jb)
	switch {
	case err != nil:
		fallthrough
	case jb.state == StateError:
		err2 = fq.moveRename(jb.name, sourceDir, "error")
	case jb.state == StateSuccess:
		err2 = fq.moveRename(jb.name, sourceDir, "success")
	}
	if err == nil && err2 != nil {
		err = err2
		err2 = nil
	}
	// since we can only return one error, log the other one
	if err2 != nil {
		log.Printf("Error moving directory %s", err2)
	}
	return err
}

// is the given job in any of the directories?
// returns the directory name, or "" if not found
func (fq *fileQueue) findJobDir(name string) string {

	for _, dir := range subdirs {
		dname := path.Join(fq.basepath, dir, name)
		_, err := os.Stat(dname)
		if err == nil {
			return dir
		}
	}
	return ""
}

// save the job to disk and close any open files inside the job structure
func (fq *fileQueue) save(jb *Job) error {
	defer func() {
		if jb.logfile != nil {
			jb.logfile.Close()
			jb.logfile = nil
		}
		if jb.log != nil {
			jb.log = nil
		}
	}()
	var err error
	jobfile, err := os.Create(path.Join(jb.path, "JOB"))
	if err != nil {
		if jb.log != nil {
			jb.log.Println("Error saving JOB file:", err)
		}
		return err
	}
	defer jobfile.Close()
	data, err := json.MarshalIndent(jb, "", "    ")
	if err == nil {
		_, err = jobfile.Write(data)
	}
	if err != nil && jb.log != nil {
		jb.log.Println("Error saving JOB file:", err)
	}
	return err
}

// Scan the queue directory and load jobs one by one until we have processed
// everything.
// If the error was related to loading a job, will return a non-nil pointer to
// a job and an error. The job is not moved to the error state.
// returns a nil Job and a nil error if there is nothing left to process.
func (fq *fileQueue) scanAndLoad() (*Job, error) {
	dentries, err := fq.listJobs("queue")
	if err != nil {
		return nil, err
	}
	for _, finfo := range dentries {
		if !finfo.IsDir() {
			continue
		}
		name := finfo.Name()
		err := fq.move(name, "queue", "processing")
		if err != nil {
			jb, _ := fq.load("queue", name)
			jb.state = StateQueue
			return jb, err // return original error
		}
		return fq.load("processing", name)
	}
	return nil, nil
}

const (
	pollDuration = 5 * time.Second
)

// watchDir watches the directory dirname and returns when either files are
// added or there is an error.
func (fq *fileQueue) watchDir(dirname string) error {
	dname := path.Join(fq.basepath, dirname)
	info, err := os.Stat(dname)
	if err != nil {
		return err
	}

	// since NFS mounts don't support ionotify events, lets poll
	for {
		// TODO: use time.tick so we don't keep allocating memory
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

// load a job named `name` from the directory `dir`.
// Will always return a stubbed out job object. May also
// return an error. If there is an error loading, that is
// passed back, but the state of the job is not changed.
// The control needs explicitly move the job into an error state.
func (fq *fileQueue) load(dir, name string) (*Job, error) {
	var err error
	jobpath := path.Join(fq.basepath, dir, name)
	result := &Job{
		name:  name,
		state: StateProcessing,
		path:  jobpath,
	}
	// open the log first, so if there is an error there is a place to put it
	result.logfile, err = os.OpenFile(
		path.Join(jobpath, "LOG"),
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0664)
	if err == nil {
		result.log = log.New(result.logfile, "", log.Ldate|log.Ltime|log.Lmicroseconds)
		err = loadJobFile(result)
	}
	// always try to load webhooks, even if there was an error.
	err2 := loadWebhooks(result)
	if err2 != nil && result.log != nil {
		result.log.Println("Error opening WEBHOOK file:", err2)
	}
	return result, err
}

// overwrites jb with the contents of its JOB file
func loadJobFile(jb *Job) error {
	jobfile, err := os.Open(path.Join(jb.path, "JOB"))
	if err != nil {
		if os.IsNotExist(err) {
			// it is okay if there is no JOB file
			return nil
		}
		jb.log.Println("Error opening JOB file:", err)
		return err
	}
	defer jobfile.Close()
	dec := json.NewDecoder(jobfile)
	err = dec.Decode(jb)
	if err != nil {
		jb.log.Println("Error reading JOB file:", err)
	}
	return err
}

// replace the webhooks in jb with those in the given "WEBHOOK" filej
func loadWebhooks(jb *Job) error {
	contents, err := ioutil.ReadFile(path.Join(jb.path, "WEBHOOK"))
	if err == nil {
		jb.webhooks = strings.Split(string(contents), "\n")
	} else if os.IsNotExist(err) {
		// don't care if the file is missing
		err = nil
	}
	return err
}

// NewFileQueue creates a new FileQueue having its state
// directories at basepath.
func NewFileQueue(basepath string) *fileQueue {
	return &fileQueue{basepath: basepath}
}

var (
	subdirs = []string{
		"queue",
		"processing",
		"success",
		"error",
		"data",
	}
)

// Initialize the given file queue's directories
func (fq *fileQueue) Init() error {
	// do the directories exist?
	for _, subdir := range subdirs {
		err := os.MkdirAll(path.Join(fq.basepath, subdir), 0744)
		if err != nil {
			return err
		}
	}

	// return any jobs in processing directory to the queue
	fnames, err := fq.listJobs("processing")
	if err != nil {
		return err
	}
	for _, finfo := range fnames {
		err = fq.move(finfo.Name(), "processing", "queue")
		if err != nil {
			return err
		}
	}
	return nil
}

// list jobs in a given status directory
func (fq *fileQueue) listJobs(dir string) ([]os.FileInfo, error) {
	sourcedir := path.Join(fq.basepath, dir)
	return ioutil.ReadDir(sourcedir)
}

// move a job between status directories
func (fq *fileQueue) move(name, srcdir, dstdir string) error {
	src := path.Join(fq.basepath, srcdir, name)
	dst := path.Join(fq.basepath, dstdir, name)
	return os.Rename(src, dst)
}

// Move a job between status directories, possibly renaming if there
// is a directory inside dstdir having the same name as `name`.
func (fq *fileQueue) moveRename(name, srcdir, dstdir string) error {
	// If we first check for the existence of a file and then
	// move the directory in two steps, there is a possibility
	// of a race condition, so we just keep trying to move
	// the directory using different source names until one works.
	var count = 1
	src := path.Join(fq.basepath, srcdir, name)
	dst := path.Join(fq.basepath, dstdir, name)
	err := os.Rename(src, dst)
	for err != nil {
		dst = path.Join(fq.basepath, dstdir, fmt.Sprintf("%s-%03d", name, count))
		count++
		err = os.Rename(src, dst)
		if count >= 1000 {
			break
		}
	}
	return err
}
