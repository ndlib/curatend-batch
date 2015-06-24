package batchs

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

type webhookMessage struct {
	Host    string `json:"host"`
	Version string `json:"version"`
	Name    string `json:"job_name"`
	State   string `json:"job_state"`
	Log     string `json:"job_log,omitempty"`
}

// Synchronously Call all the webhooks with the current task information
func (ctx *Context) callWebhooks(jb *Job) error {
	if len(jb.webhooks) == 0 {
		return nil
	}
	m := webhookMessage{
		Version: ctx.version,
		Name:    jb.name,
		State:   jb.state.String(),
	}
	m.Host, _ = os.Hostname()
	if jb.state == StateError {
		var err error
		m.Log, err = ctx.logfileForJob(jb)
		if err != nil {
			log.Println("webhook:", err)
		}
	}
	buffer, err := json.Marshal(m)
	if err != nil {
		log.Println("callWebhooks:", err)
		return err
	}
	reader := bytes.NewReader(buffer)
	for _, url := range jb.webhooks {
		if url == "" {
			continue
		}
		if !strings.HasPrefix(url, "http") {
			url = "http://" + url
		}
		reader.Seek(0, 0)
		r, err := http.Post(url, "application/json", reader)
		if err != nil {
			log.Printf("callWebhooks: %s: %s: %s\n", jb.name, url, err.Error())
			if jb.log != nil {
				jb.log.Printf("Webhook: %s: %s\n", url, err.Error())
			}
		} else {
			// don't care about response, but we need to close it
			r.Body.Close()
		}
	}
	return nil
}

// Return the complete contents of the log file associated to the given job.
// This routine goes around the abstraction of the Job structure,
// and assumes it is implemented using a directory containing a file
// named "LOG". A better way to support this functionality should be found.
func (ctx *Context) logfileForJob(jb *Job) (string, error) {
	logfile := path.Join(jb.path, "LOG")
	contents, err := ioutil.ReadFile(logfile)
	return string(contents), err
}
