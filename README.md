# CurateND-Batch

This repository provides code to run the batch ingest infrastructure for
[CurateND].  Most of the code to perform the actual processing is in the
[ROF] gem.  This code provides a simple job manager.  It is a work in
progress, and we may move to a standard work-queue system in the future.

[CurateND]: https://curate.nd.edu
[ROF]: https://github.com/ndlib/rof

# Format of this Repository

This repository is deployed using capistrano and puppet, and, since it is
"rubyish", it also invokes bundler. The deploy then compiles the Go
code, which is in a subdirectory, and installs that.  Finally the puppet
scripts run and setup the environment, and install the configuration files
(e.g. start the Go service at startup).

The Go service will monitor a directory for new files. It will spawn a
ROF process to handle the new files.  In this way most of the processing
logic is in the ROF code, and the Go code is only concerned with monitoring
directories and starting processes.  The Go code **does** track workflow
steps, and provides a REST API to monitor and control the overall process.
