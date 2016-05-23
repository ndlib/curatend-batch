#!/bin/bash

# Start and configure the curatend-batch service.
# This script is only intended for use in a development environment.

BATCHDIR="$HOME/goapps/curatend-batch"

# There can be multiple environments within the batch space
mkdir -p "$BATCHDIR/development"
curatend-batch -queue="$BATCHDIR/development" -tasks="$GOPATH/src/github.com/ndlib/curatend-batch/tasks"
