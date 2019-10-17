#!/usr/bin/env ruby
# frozen_string_literal: true

# This script wi8ll look for a subfolder in the current JOBPATH
# named TO_IIIF. If it exists, it will iterate throught the xml files
# in that direction, and upload each into a folder of the same name
# in the IIIF pipeline google drive (MARBLE_INGEST_DRIVE_ID) appropriate
# for this environment

require 'google_drive'
require 'csv'

# JOBPATH will be set by batch ingest.
jobpath = ENV['JOBPATH']

# if TO_IIIF exists, we have work to do. If Not. we're done

unless File.directory?("#{jobpath}/TO_IIIF")
  p 'No works to upload to IIIF pipeline.'
  works 0
end

# check that GOOGLE_API_JSON is set- path of credentials file
if ENV['GOOGLE_API_JSON'].nil?
  abort('ERROR: ENV variable GOOGLE_API_JSON must be set to location of credentials json file')
end

# check that GOOGLE_API_JSON is set- path of credentials file
if ENV['MARBLE_INGEST_DRIVE_ID'].nil?
  abort('ERROR: ENV variable MARBLE_INGEST_DRIVE_ID must be set to the googel drive id of the marble ingest folder')
end

# set the work pid, credentials file, and drive_id
service_credentials = ENV['GOOGLE_API_JSON']
drive_id = ENV['MARBLE_INGEST_DRIVE_ID']

# Creates a session using a service account, and uploads files to google drive
# https://github.com/gimite/google-drive-ruby/blob/master/doc/authorization.md

unc_service_credentials = Base64.strict_decode64(service_credentials)
stringy_json_creds = StringIO.new(unc_service_credentials)
session = GoogleDrive::Session.from_service_account_key(stringy_json_creds)

# get environment-specifer folder
curate_drive = session.file_by_url("https://drive.google.com/drive/folders/#{drive_id}")

Dir.chdir("#{jobpath}/TO_IIIF")

# Iterate thru the works we are uploading
# if the work is new, create a subfolder for it
# if the work already exists, delete the work xml
# and any other files we are uploading first, then replace them
# with the updates

Dir.glob("#{jobpath}/TO_IIIF/*") do |work|
  work_pid = File.basename(work, File.extname(work))
  # Does work dir exist?
  works_folder = nil
  this_work_dir = curate_drive.files(include_team_drive_items: true, q: ['trashed = false AND name = ?', work_pid] )

  # if new work, create subfolder else reuse old one,first deleting its contents
  if this_work_dir.length == 1
    p "#{work_pid} already exists in google drive - overwriting"
    works_folder = this_work_dir[0]
    works_folder.files(include_team_drive_items: true, q: 'trashed = false') do |file|
      file.delete(permanent: true)
    end
  else
    p "#{work_pid} does not exist in google drive - creating"
    works_folder = curate_drive.create_subfolder(work_pid)
  end
  works_folder.upload_from_file("#{work_pid}/#{work_pid}.xml", "#{work_pid}.xml")

  #If sequence.csv exists, use it as list to upload files from this work
  
  # the files associated with is work are in the Filenames column of sequence.csv
  if File.exists?("#{jobpath}/TO_IIIF/#{work_pid}/sequence.csv")
    sequence_table = CSV.parse(File.read("#{jobpath}/TO_IIIF/#{work_pid}/sequence.csv"), quote_char: "'", headers: true)

    #For each file- delete it if it altready exists then upload to Google_ENV_Folder/work_pid/file_name
    (0).upto(sequence_table.length-1) do |row|
      file_name = sequence_table[row]['Filenames']
      works_folder.files(include_team_drive_items: true, q: ['trashed = false AND name = ?', file_name]) do |file|
        file.delete(permanent: true)
      end
      local_copy = "#{jobpath}/TO_TAPE/#{work_pid}/#{file_name}" 
      p "Uploading #{local_copy}"
      works_folder.upload_from_file( local_copy, file_name)
    end
  end

end
