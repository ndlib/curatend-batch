#!/usr/bin/env ruby
#
# A script to help remediate sipity jobs that made it all the way to ingest
# task before failing. It makes a copy of the failed job in the data directory
# of the target environment, cleans up the JOB and LOG from the old run, and moves
# the datafiles back to the job rootdir (DLTP-1125)
require 'fileutils'
require 'time'

#check for 2 arguments
abort('Usage: requeue-sipity-job environment <failed-job-directory>') if ARGV.length != 2

#  verify that second argument is a directory
abort("#{ARGV[1]} is not a directory") if !File.directory?(ARGV[1])
 
#verify environment
case ARGV[0]
  when 'production'
    batch_rootdir='/mnt/curatend-batch/production'
  when 'pre_production'
    batch_rootdir='/mnt/curatend-batch/test/pprd'
  when 'staging'
    batch_rootdir='/mnt/curatend-batch/test/libvirt9'
  else
    abort("Unknown environment #{ARGV[0]}")
end
 
#copy failed job to data directory
sipity_job=File.basename(ARGV[1])
data_dir=File.join(batch_rootdir,'data')
sipity_dir=File.join(data_dir,sipity_job)

FileUtils.cp_r(ARGV[1], data_dir, :verbose => true )

puts "Moved #{ARGV[1]} to #{sipity_dir}"

#remove sipity job's old run state, logs, and callback
File.delete(File.join(sipity_dir, 'LOG'))
File.delete(File.join(sipity_dir, 'JOB'))
File.delete(File.join(sipity_dir, 'WEBHOOK'))

#move the datafiles that were set up for bendo in TO_TAPE back to the job root directory
Dir.foreach(File.join(sipity_dir, 'TO_TAPE')) { |bendo_pid_dir|
  next if bendo_pid_dir == '.' || bendo_pid_dir == '..'
  full_pid_dir_path = File.join(sipity_dir, 'TO_TAPE', bendo_pid_dir)
  FileUtils.rm_r(File.join(full_pid_dir_path, 'fedora3'))
  Dir.foreach(full_pid_dir_path) { |data_file|
    next if data_file == '.' || data_file == '..'
    FileUtils.mv(File.join(full_pid_dir_path,data_file), File.join(sipity_dir, data_file))
  }
}

# remove old TO_TAPE directory
FileUtils.rm_r(File.join(sipity_dir, 'TO_TAPE'))
