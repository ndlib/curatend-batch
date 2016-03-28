package batchs

import (
	"encoding/json"
	"fmt"
	"io"
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
			w.WriteHeader(404)
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

	if dir != "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		if dir != "data" {
			response.Name = id
			response.Status = dir
		} else {
			response.Name = id
			response.Status = "ready"
		}
		enc.Encode(response)
		return
	}

	w.WriteHeader(404)
	fmt.Fprintln(w, "No Job Id Found")
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

	err2 := os.Mkdir(jobPath, 0744)

	if err2 != nil {
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

	w.WriteHeader(200)
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
		jobPath := path.Join(s.QueuePath.basepath, dir, id)

		if _, err := os.Stat(jobPath); err == nil {

			err2 := s.QueuePath.move(id, dir, "queue")

			if err2 != nil {
				w.WriteHeader(500)
				fmt.Fprintln(w, err.Error())
			}

			w.WriteHeader(200)
			return
		}
	default:
		w.WriteHeader(404)
		fmt.Fprintln(w, "Job Id Not Found")
	}

}

// PutJobIdFileHandler implements PUT /jobs/:id/files/*path
// Returns 404 if data directory for given job id does not exist

func (s *RESTServer) PutJobIdFileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")
	filePath := path.Clean(ps.ByName("path"))

	if filePath == "." {
		w.WriteHeader(400)
		fmt.Fprintln(w, "Non-Lexical File Path ")
		return
	}

	var err error

	dir := s.QueuePath.findJobDir(id)
	switch dir {
	case "queue", "processing", "":
		w.WriteHeader(404)
		fmt.Fprintln(w, err.Error())
		return
	}

	// from here on, fullUploadPath is the file target destination

	fullUploadPath := path.Join(s.QueuePath.basepath, dir, filePath)

	// if there's no body, we've got nothing to upload
	if r.Body == nil {
		w.WriteHeader(400)
		fmt.Fprintln(w, "no body")
		return
	}

	// ensure that the directory path to the file is present

	err = os.MkdirAll(path.Dir(fullUploadPath), 0774)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err.Error())
		return
	}

	// open target file- if it already exists, truncate and overwrite
	fileInfo, err := os.OpenFile(fullUploadPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)

	defer fileInfo.Close()
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err.Error())
		return
	}

	_, err = io.Copy(fileInfo, r.Body)
	w.WriteHeader(200)
}

// DeleteJobIdFileHandler implements DELETE /jobs/:id/files/*path
// Returns 404 if data directory for given job id does not exist

func (s *RESTServer) DeleteJobIdFileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")
	filePath := path.Clean(ps.ByName("path"))

	if filePath == "." {
		w.WriteHeader(400)
		fmt.Fprintln(w, "Non-Lexical File Path ")
		return
	}

	var err error

	dir := s.QueuePath.findJobDir(id)
	switch dir {
	case "queue", "processing", "":
		w.WriteHeader(404)
		fmt.Fprintln(w, err.Error())
		return
	}

	// from here on, fullUploadPath is the file target destination

	fullDeletePath := path.Join(s.QueuePath.basepath, dir, filePath)

	err = os.Remove(fullDeletePath)

	if err != nil {
		fmt.Fprintln(w, err.Error())
	}
}

// GetJobIdFileHandler implements GET /jobs/:id/files/*path
// Returns 404 if job directory, or desired file for given job, does not exist

func (s *RESTServer) GetJobIdFileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")
	filePath := ps.ByName("path")

	var err error

	dir := s.QueuePath.findJobDir(id)
	switch dir {
	case "queue", "processing", "":
		w.WriteHeader(404)
		fmt.Fprintln(w, err.Error())
		return
	}

	// from here on, fullDownloadPath is the file target destination

	fullDownloadPath := path.Join(s.QueuePath.basepath, dir, filePath)

	// if the target file does not exist, return Not Found
	if _, err = os.Stat(fullDownloadPath); os.IsNotExist(err) {
		w.WriteHeader(404)
		fmt.Fprintln(w, err.Error())
		return
	}

	fileInfo, err := os.OpenFile(fullDownloadPath, os.O_RDONLY, 0664)

	defer fileInfo.Close()

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	_, err = io.Copy(w, fileInfo)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err.Error())
		return
	}
}
