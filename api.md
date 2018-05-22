# Batch Ingest API


This document describes an new HTTP based API to the batch ingest service. This
API is not intended to replace the file-system based API the batch ingest
currently uses. This API is intended to provide a convenient, non-file system
based, way for other services to create and check on batch jobs. It is intended
to be used for jobs with only a modest number of files and modestly sized
files.

## Preliminaries

The HTTP API is by default on port XXXXX. The basic unit is the job. All
requests require a token in the `X-Api-Token` header. The `Accept-Type` header
is ignored--all responses are sent as `application/json`.

## Routes

See [batchs/routes.go](./batchs/routes.go) for API exposed routes.

### List all jobs

`GET /jobs`

This returns a JSON array listing the names of all the jobs, regardless of
their status. It probably should have pagination, but doesn't. Example:

    GET /jobs

```json
[
  { "Name": "job-1", "Status": "success" },
  { "Name": "another-job", "Status": "success" },
  { "Name": "test2", "Status": "success" }
]
```

### Get Job Info

`GET /jobs/:jobid`

This request returns the job's information as a JSON object. Example:

`GET /jobs/test2`

```json
{
    "Name": "test2",
    "Status": "success"
}
```

The possible statuses are:

 * "success" - the job completed processing successfully,
 * "error", - the job had an error during processing,
 * "processing", - the job is being processed,
 * "queue", - the job is waiting to be processed,
 * or "ready" - the job has not been queued yet.

### Create of Update Job Info

`PUT /jobs/:jobid`

A payload does not need to be provided. If one is present, it is ignored. This
call is mainly used to create a job if a job with that id does not already
exist.

### Delete a job

`DELETE /jobs/:jobid`

This will delete the given job and all the data and files inside it.

### Submit a job

`POST /jobs/:jobid/queue`

This request will ask that the given job be moved to the `queue` directory. It
has no effect if the job is already in the `queue` or `processing` directories.
If a job is moved from the "error" or "success" states to "queue", it will
resume with whatever task caused an error, or the next task in the Todo list.
To reprocess a job from the complete beginning `DELETE` the `JOB` file and then
submit the job.

### Add a file

`PUT /jobs/:jobid/files/*path`

This request will add a file into the given job. The path given may contain
forward slashes, and those will create subdirectories. It is legal to alter
files used by the batch ingest service, such as `JOB`, `LOG`, and `WEBHOOK`.
The request body is saved as the contents of the file. If the file already
exists, it is replaced.

### Get a file

`GET /jobs/:jobid/files/*path`

This request will return the contents of a file inside of the given job. If the
`*path` is a directory, the response will be an Array of filenames and
subdirectories within the directory. (see "Get a list of files" below)

### Get a list of files

`GET /jobs/:jobid/files`

This request will return an Array of filenames and subdirectories.

```json
["filename", "subdirectory"]
```

### Remove a file

`DELETE /jobs/:jobid/files/*path`

This request will delete a file from the given job. It is not an error to
remove a file which does not exist.

## Implementation Notes

This API is designed to be very simple, and to map as much as possible directly
to the original file-system based API. The one key change is an additional
`data` directory to hold the jobs which were created for this batch upload
service but haven't been run yet.
