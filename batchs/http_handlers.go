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

// ItemHandler handles requests to GET /jobs

func (s *RESTServer) GetJobsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var jobs []string

	// subdirs contains a list of all the batchs directories

	for _, dir := range subdirs {

		// list jobs in each directory, adding them to jobs array

		inThisDir, err := s.QueuePath.listJobs(dir)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err.Error())
			return
		}

		for _, thisJob := range inThisDir {
			jobs = append(jobs, thisJob.Name())
		}
	}

	// convert jobs to JSON- that is our returned content
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.Encode(jobs)
}

// GetJobIdHandler handles requests to GET /jobs/:id
// Returns 404 if job id not found
//         200 + Json { "Name": id, "Status": [ success, error, queue, processing, ready ] }

func (s *RESTServer) GetJobIdHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var response struct{ Name, Status string }

	id := ps.ByName("id")

	dir := s.QueuePath.findJobDir(id)

	if dir == "" {
		w.WriteHeader(404)
		fmt.Fprintln(w, "Not Found")
		return
	}
	response.Name = id
	if dir == "data" {
		response.Status = "ready"
	} else {
		response.Status = dir
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.Encode(response)
}

// PutJobIdHandler handles requests to PUT /jobs/:id
// Returns 200 if id directory can be created under data
//         403 if it cannot (already exists)

func (s *RESTServer) PutJobIdHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")

	jobPath := path.Join(s.QueuePath.basepath, "data", id)

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

// DeleteJobIdHandler handles requests to DELETE /jobs/:id
// Returns 200 if id directory was deleted in one or more of [success, error, data], or doesn't exist
//         500 if something went wrong

func (s *RESTServer) DeleteJobIdHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")

	// Remove job id wherever it's found

	dir := s.QueuePath.findJobDir(id)

	if dir != "" {
		jobPath := path.Join(s.QueuePath.basepath, dir, id)

		err := os.RemoveAll(jobPath)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err.Error())
		}
	}
}

// SubmitJobIdHandler handles requests to POST /jobs/:id
// Returns 200 if id directory was queued from one  of [data, error, success],  in that order
//         500 if something went wrong
//         404 if the job id was not found  in any of the searched directories

func (s *RESTServer) SubmitJobIdHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")

	dir := s.QueuePath.findJobDir(id)

	switch dir {
	case "data", "error", "success":
		err := s.QueuePath.move(id, dir, "queue")
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

// PutJobIdFileHandler implements PUT /jobs/:id/files/*path
// Returns 404 if data directory for given job id does not exist

func (s *RESTServer) PutJobIdFileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")
	filepath := ps.ByName("path")
	// from here on, fullUploadPath is the file target destination
	fullUploadPath := s.findFile(w, id, filepath)
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

// DeleteJobIdFileHandler implements DELETE /jobs/:id/files/*path
// Returns 404 if data directory for given job id does not exist

func (s *RESTServer) DeleteJobIdFileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")
	filepath := ps.ByName("path")
	// from here on, fullUploadPath is the file target destination
	fullDeletePath := s.findFile(w, id, filepath)
	if fullDeletePath == "" {
		return
	}

	err := os.Remove(fullDeletePath)

	if err != nil {
		fmt.Fprintln(w, err.Error())
	}
}

// GetJobIdFileHandler implements GET /jobs/:id/files/*path
// Returns 404 if job directory, or desired file for given job, does not exist

func (s *RESTServer) GetJobIdFileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")
	filepath := ps.ByName("path")
	// from here on, fullDownloadPath is the file target destination
	fullDownloadPath := s.findFile(w, id, filepath)
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
		var result []string
		for _, dent := range list {
			result = append(result, dent.Name())
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
	io.Copy(w, fileInfo)
}

// findFile takes a job id and a filepath into that job. It will sanitize the
// filepath, and then try to resolve the path to an absolute path. It will
// return errors directly to w as found. It does not allow any processing on
// jobs which are in the queue or processing directory. An empty string
// is returned if an error message is written to w.
func (s *RESTServer) findFile(w http.ResponseWriter, id, filepath string) string {
	filepath = path.Clean(filepath)

	if filepath == "." {
		w.WriteHeader(400)
		fmt.Fprintln(w, "Non-Lexical File Path ")
		return ""
	}

	dir := s.QueuePath.findJobDir(id)
	switch dir {
	case "queue", "processing":
		w.WriteHeader(409)
		fmt.Fprintf(w, "Cannot access queued jobs")
		return ""
	case "":
		w.WriteHeader(404)
		return ""
	default:
		return path.Join(s.QueuePath.basepath, dir, id, filepath)
	}
}
