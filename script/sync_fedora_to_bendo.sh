#!/bin/bash

# Try to get a lock, and exit if someone else already has it.
# This keeps a sync processes from spawning nd overlapping
# should a paricular sync take a long time.
# The lock is released when this shell exits.

cd /home/app/curatend-batch/current

exec 200> "/tmp/sync_fedora_to_bendo"
flock -e --nonblock 200 || exit 0

# source our ruby env
source /etc/profile.d/ruby.sh

# source our app environment.

#source /home/app/curatend-batch/shared/system/metrics-env.sh

/opt/ruby/current/bin/bundle exec rake batchs:make_bendo_updatelist
