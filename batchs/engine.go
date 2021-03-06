package batchs

import (
	"log"
	"time"
)

// Context tracks the path of where we store the queues.
// It is the only structure which knows about how the queues are persisted
// to disk.
// It should be promoted to an interface, probably.
type Context struct {
	taskpath string    // the path to the directory mapping task names to commands
	q        JobSource // the source of new jobs to run
	version  string    // the version of this server (passed in)
}

// Run the main event loop forever. Errors are logged.
// This function never returns.
func (ctx *Context) Run() error {
	err := ctx.q.Init()
	if err != nil {
		return err
	}
	for {
		var runjob = true
		jb, err := ctx.q.NextJob()
		if err != nil {
			log.Println("Run:", err)
			runjob = false
		}
		if jb == nil {
			log.Println("Run: No job, sleeping for a bit")
			time.Sleep(1 * time.Second)
			continue
		}
		log.Printf("Processing job %s", jb.name)
		// run webhook for "processing" state
		ctx.callWebhooks(jb)
		// setup the job structure
		jb.taskpath = ctx.taskpath
		jb.state = StateError // assume there was an error
		if len(jb.Todo)+len(jb.Finished) == 0 {
			// If a job has no tasks assigned and has completed no tasks
			// seed it with the default task `start`
			jb.Todo = append(jb.Todo, "start")
		}
		starttime := time.Now()
		if runjob {
			if err = jb.process(); err == nil {
				jb.state = StateSuccess
			}
		}
		// Record the duration of the job to be scraped out later.
		// Don't change the format of the log line without speaking to ESU
		duration := time.Now().Sub(starttime)
		log.Printf("Done processing job %s (%s) %.3fs", jb.name, jb.state, duration.Seconds())
		ctx.callWebhooks(jb)
		err = ctx.q.FinishJob(jb)
		if err != nil {
			log.Println("FinishJob:", err)
		}
	}
}

// NewContext creates a new context structure.
// taskpath is used to resolve task names into commands.
func NewContext(q JobSource, taskpath, version string) *Context {
	return &Context{
		q:        q,
		taskpath: taskpath,
		version:  version,
	}
}
