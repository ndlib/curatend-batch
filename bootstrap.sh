#!/bin/bash

# Start and configure the curatend-batch service.
# This script is only intended for use in a development environment.

BATCHDIR="$HOME/goapps/curatend-batch"

# There can be multiple environments within the batch space
mkdir -p "$BATCHDIR/development"

# Setting up a config file
config_file=tasks/conf
if [ -f "$config_file" ]; then
  echo "Config file already exists"
else
  echo "Creating a default config file"

  # Escaping expanded $GOPATH
  target=$(echo "$GOPATH" | awk '{gsub("/", "\\\\&");print}')

  sed s/=\$GOPATH/="$target"/ "$config_file".example > "$config_file"
fi

curatend-batch -queue="$BATCHDIR/development" -tasks="$GOPATH/src/github.com/ndlib/curatend-batch/tasks"
