# CurateND-Batch

[![APACHE 2
License](http://img.shields.io/badge/APACHE2-license-blue.svg)](./LICENSE)
[![Contributing
Guidelines](http://img.shields.io/badge/CONTRIBUTING-Guidelines-blue.svg)](./CONTRIBUTING.md)
[![Go Report
Card](https://goreportcard.com/badge/github.com/ndlib/curatend-batch)](https://goreportcard.com/report/github.com/ndlib/curatend-batch)
<img src="http://online.swagger.io/validator?url=https://raw.githubusercontent.com/ndlib/curatend-batch/master/definitions/swagger.yml">

This repository provides code to run the batch ingest infrastructure for
[CurateND].  Most of the code to perform the actual processing is in the
[ROF] gem.  This code provides a simple job manager.  It is a work in
progress, and we may move to a standard work-queue system in the future.

[CurateND]: https://curate.nd.edu
[ROF]: https://github.com/ndlib/rof

# Format of this Repository

This repository is deployed using capistrano and ansible, and, since it is
"rubyish", it also invokes bundler. The Go
code, which is in a subdirectory, is compiled and deployed separately usiong AWS codebuild.
The ansible scripts run and setup the environment, and install the configuration files
(e.g. start the Go service at startup).

The Go service will monitor a directory for new files. It will spawn a
ROF process to handle the new files.  In this way most of the processing
logic is in the ROF code, and the Go code is only concerned with monitoring
directories and starting processes.  The Go code **does** track workflow
steps, and provides a REST API to monitor and control the overall process.

# Getting Started

To run locally:

* Make sure you have `$GOROOT` and `$GOPATH` set.
* `cd /path/to/this/repository`
* `go build` - This will make an executable file
* `./curatend-batch` - Run the executable; it will by default run on port 15000

## Running tests

For batch:

```console
$ cd batches
$ go test
```

## Sentry

CurateND-Batch can optionally send error messages to the Sentry service. Enable it by setting
the environment variables `SENTRY_DSN`, `SENTRY_RELEASE`, and `SENTRY_ENVIRONMENT`.

# Documentation

* Detailed documentation is available in a [Google Document](https://docs.google.com/document/d/1LtI4QlowI-rcJdFy3535T8417NKPU4FW-zpGi-b-XYk/edit) (ND only).
* [API Documentation](./api.md)
