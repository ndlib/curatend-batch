package batchs

import (
	"encoding/json"
	"errors"
	"fmt"
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

	var response map[string]string

	id := ps.ByName("id")

	for _, dir := range subdirs {

		filePath := path.Join(s.QueuePath.basepath, dir, id)
		_, err := os.Stat(filePath)

		if err == nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			enc := json.NewEncoder(w)
			if dir != "data" {
				response = map[string]string{"Name": id, "Status": dir}
			} else {
				response = map[string]string{"Name": id, "Status": "ready"}
			}
			enc.Encode(response)
			return
		}
	}

	w.WriteHeader(404)
	fmt.Fprintln(w, errors.New("No Job Id Found"))
}

// PutJobIdHandler handles requests to PUT /jobs/:id
// Returns 200 if id directory can be created under data
//         403 if it cannot (already exists)

func (s *RESTServer) PutJobIdHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")

	jobPath := path.Join(s.QueuePath.basepath, "data", id)

	if _, err := os.Stat(jobPath); err == nil {
		w.WriteHeader(403)
		fmt.Fprintln(w, errors.New("Job Already Exists"))
		return
	}

	err2 := os.Mkdir(jobPath, 0744)

	if err2 != nil {
		w.WriteHeader(403)
		fmt.Fprintln(w, errors.New("Error Creating Job  "))
	}
}

// DeleteJobIdHandler handles requests to DELETE /jobs/:id
// Returns 200 if id directory was deleted in one or more of [success, error, data], or doesn't exist
//         500 if something went wrong

func (s *RESTServer) DeleteJobIdHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var jobdirs = []string{
		"success",
		"error",
		"data",
	}

	id := ps.ByName("id")

	// Remove job id wherever it's found

	for _, dir := range jobdirs {

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

	var jobdirs = []string{
		"data",
		"error",
		"success",
	}

	id := ps.ByName("id")

	// Remove job id wherever it's found

	for _, dir := range jobdirs {

		jobPath := path.Join(s.QueuePath.basepath, dir, id)

		if _, err := os.Stat(jobPath); err == nil {

			err2 := s.QueuePath.move(id, dir, "queue")

			if err2 != nil {
				w.WriteHeader(500)
				fmt.Fprintln(w, err.Error())
			}

			return
		}
	}

	// if we got this far, the job id given did not exist
	w.WriteHeader(404)
	fmt.Fprintln(w, errors.New("Job Id Not Found"))
}
