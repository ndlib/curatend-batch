package batchs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/julienschmidt/httprouter"
)

// WelcomeHandler writes the version number as a response.
func (server *RESTServer) WelcomeHandler(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// if it's a GET, copy the data into the response- if it's a HEAD, don't
	if request.Method == "HEAD" {
		return
	}
	fmt.Fprintf(writer, "CurateND Batch (%s)\n", server.Version)
}

// GetJobsHandler handles requests to GET /jobs
func (server *RESTServer) GetJobsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var jobs []JobInfo

	// subdirs contains a list of all the batchs directories

	for _, dir := range subdirs {

		// list jobs in each directory, adding them to jobs array

		inThisDir, err := server.QueuePath.listJobs(dir)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err.Error())
			return
		}

		for _, thisJob := range inThisDir {
			s := JobInfo{
				Name:   thisJob.Name(),
				Status: dir,
			}
			if dir == "data" {
				s.Status = "ready"
			}
			jobs = append(jobs, s)
		}
	}

	// convert jobs to JSON- that is our returned content
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.Encode(jobs)
}

// GetJobIDHandler handles requests to GET /jobs/:id
// Returns 404 if job id not found
//         200 + Json { "Name": id, "Status": [ success, error, queue, processing, ready ] }
func (server *RESTServer) GetJobIDHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var response JobInfo

	id := ps.ByName("id")

	dir := server.QueuePath.findJobDir(id)

	if dir == "" {
		w.WriteHeader(404)
		fmt.Fprintln(w, "Not Found")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// if it's a GET, copy the data into the response- if it's a HEAD, don't
	if r.Method == "HEAD" {
		return
	}

	response.Name = id
	if dir == "data" {
		response.Status = "ready"
	} else {
		response.Status = dir
	}
	enc := json.NewEncoder(w)
	enc.Encode(response)
}

// A JobInfo contains the information on jobs that we send to clients.
type JobInfo struct {
	Name, Status string
}

// PutJobIDHandler handles requests to PUT /jobs/:id
// Returns 200 if id directory can be created under data
//         403 if it cannot (already exists)
func (server *RESTServer) PutJobIDHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")

	jobPath := path.Join(server.QueuePath.basepath, "data", id)

	if _, err := os.Stat(jobPath); err == nil {
		w.WriteHeader(403)
		fmt.Fprintln(w, "Job Already Exists")
		return
	}

	err := os.Mkdir(jobPath, 0744)

	if err != nil {
		w.WriteHeader(403)
		fmt.Fprintln(w, "Error Creating Job")
	}
}

// DeleteJobIDHandler handles requests to DELETE /jobs/:id
// Returns 200 if id directory was deleted in one or more of [success, error, data], or doesn't exist
//         500 if something went wrong
func (server *RESTServer) DeleteJobIDHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")

	// Remove job id wherever it's found

	dir := server.QueuePath.findJobDir(id)

	if dir != "" {
		jobPath := path.Join(server.QueuePath.basepath, dir, id)

		err := os.RemoveAll(jobPath)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err.Error())
		}
	}
}

// SubmitJobIDHandler handles requests to POST /jobs/:id
// Returns 200 if id directory was queued from one  of [data, error, success],  in that order
//         500 if something went wrong
//         404 if the job id was not found  in any of the searched directories
func (server *RESTServer) SubmitJobIDHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")

	dir := server.QueuePath.findJobDir(id)

	switch dir {
	case "data", "error", "success":
		err := server.QueuePath.move(id, dir, "queue")
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err.Error())
			return
		}
	case "queue", "processing":
		w.WriteHeader(409)
		fmt.Fprintln(w, "Cannot submit queued job")
	default:
		w.WriteHeader(404)
		fmt.Fprintln(w, "Not Found")
	}

}

// PutJobIDFileHandler implements PUT /jobs/:id/files/*path
// Returns 404 if data directory for given job id does not exist
func (server *RESTServer) PutJobIDFileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")
	filepath := ps.ByName("path")
	// from here on, fullUploadPath is the file target destination
	fullUploadPath := server.findFile(w, id, filepath)
	if fullUploadPath == "" {
		return
	}

	// ensure that the directory path to the file is present

	err := os.MkdirAll(path.Dir(fullUploadPath), 0774)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err.Error())
		return
	}

	// open target file- if it already exists, truncate and overwrite
	fileInfo, err := os.OpenFile(fullUploadPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err.Error())
		return
	}
	defer fileInfo.Close()

	_, err = io.Copy(fileInfo, r.Body)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err.Error())
		return
	}
}

// DeleteJobIDFileHandler implements DELETE /jobs/:id/files/*path
// Returns 404 if data directory for given job id does not exist
func (server *RESTServer) DeleteJobIDFileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")
	filepath := ps.ByName("path")
	// from here on, fullUploadPath is the file target destination
	fullDeletePath := server.findFile(w, id, filepath)
	if fullDeletePath == "" {
		return
	}

	err := os.Remove(fullDeletePath)

	if err != nil {
		fmt.Fprintln(w, err.Error())
	}
}

// GetJobIDFileHandler implements GET /jobs/:id/files/*path
// Returns 404 if job directory, or desired file for given job, does not exist
func (server *RESTServer) GetJobIDFileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")
	filepath := ps.ByName("path")
	// from here on, fullDownloadPath is the file target destination
	fullDownloadPath := server.findFile(w, id, filepath)
	if fullDownloadPath == "" {
		return
	}

	// if the target file does not exist, return Not Found
	fs, err := os.Stat(fullDownloadPath)
	if os.IsNotExist(err) {
		w.WriteHeader(404)
		return
	} else if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err.Error())
		return
	}
	if fs.IsDir() {
		list, err := ioutil.ReadDir(fullDownloadPath)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		// if it's a GET, copy the data into the response- if it's a HEAD, don't
		if r.Method == "HEAD" {
			return
		}
		var result []string
		for _, dent := range list {
			result = append(result, dent.Name())
		}
		enc := json.NewEncoder(w)
		enc.Encode(result)
		return
	}

	fileInfo, err := os.OpenFile(fullDownloadPath, os.O_RDONLY, 0664)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err.Error())
		return
	}
	defer fileInfo.Close()
	// if it's a GET, copy the data into the response- if it's a HEAD, don't
	if r.Method == "HEAD" {
		return
	}
	io.Copy(w, fileInfo)
}

// findFile takes a job id and a filepath into that job. It will sanitize the
// filepath, and then try to resolve the path to an absolute path. It will
// return errors directly to w as found. It does not allow any processing on
// jobs which are in the queue or processing directory. An empty string
// is returned if an error message is written to w.
func (server *RESTServer) findFile(w http.ResponseWriter, id, filepath string) string {
	filepath = path.Clean(filepath)

	if filepath == "." {
		w.WriteHeader(400)
		fmt.Fprintln(w, "Non-Lexical File Path ")
		return ""
	}

	dir := server.QueuePath.findJobDir(id)
	switch dir {
	case "queue", "processing":
		w.WriteHeader(409)
		fmt.Fprintf(w, "Cannot access queued jobs")
		return ""
	case "":
		w.WriteHeader(404)
		return ""
	default:
		return path.Join(server.QueuePath.basepath, dir, id, filepath)
	}
}
