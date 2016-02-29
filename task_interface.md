Task Interface
==============

Each job consists of a list of tasks.
A task is an executable file inside the batch daemon's task directory.
Tasks are run in a child process with both `STDOUT` and `STDERR` redirected into the current job's `LOG` file.
Tasks are expected to exit with a 0 status code if everything was successful and a non-zero code if there was an error.
If there was an error, the job is moved to the error directory, and upon resumption the current task will be run again.
Tasks should be able to handle a partial success and clean-up or restart appropriately.

# Task Environment

When a task is started, its current directory is the directory containing the current job to process (i.e. the `JOBPATH` directory).
Both stderr and stdout will be captured, but not interpreted, by the batch daemon.
Tasks are run as the same user as the batch daemon.
There are three defined environmental variables.

 * `JOBPATH`, the path to the job directory this task should process
 * `JOBNAME`, the id of this particular job
 * `JOBCONTROL`, a temporary file the task can write to to alter the processing of this job

The `JOBCONTROL` is a file which is read by the batch daemon after the task exits. It consists of a sequence of control messages, with one per line.
The file is processed line-by-line and processing is aborted on error.
There is only one command at the moment:

 * `addtask:`. A line beginning with `addtask:` will append the rest of the line to the end of the current job's Todo list. Example:

        addtask:assign-pids
        addtask:calculate-checksums

    This will append the tasks "assign-pids" and then "calculate-checksums" (in this order) to the end of the job's Todo list.

