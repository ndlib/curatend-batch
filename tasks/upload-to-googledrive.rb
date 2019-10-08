#!/usr/bin/env ruby

# This script wi8ll look for a subfolder in the current JOBPATH
# named TO_IIIF. If it exists, it will iterate throught the xml files
# in that direction, and upload each into a folder of the same name 
# in the IIIF pipeline google drive (MARBLE_INGEST_DRIVE_ID) appropriate
# for this environment 


require "google_drive"

# JOBPATH will be set by batch ingest. 
jobpath = ENV['JOBPATH']

# if TO_IIIF exists, we have work to do. If Not. we're done

if !File.directory?("#{jobpath}/TO_IIIF")
  p "No works to upload to IIIF pipeline."
  works 0
end

#check that GOOGLE_API_JSON is set- path of credentials file
if ENV['GOOGLE_API_JSON'].nil?
	abort("ERROR: ENV variable GOOGLE_API_JSON must be set to location of credentials json file")
end

#check that GOOGLE_API_JSON is set- path of credentials file
if ENV['MARBLE_INGEST_DRIVE_ID'].nil?
	abort("ERROR: ENV variable MARBLE_INGEST_DRIVE_ID must be set to the googel drive id of the marble ingest folder")
end

# set the work pid, credentials file, and drive_id
service_credentials = ENV['GOOGLE_API_JSON']
drive_id = ENV['MARBLE_INGEST_DRIVE_ID']

# Creates a session using a service account, and uploads files to google drive
# https://github.com/gimite/google-drive-ruby/blob/master/doc/authorization.md

session = GoogleDrive::Session.from_service_account_key(service_credentials)

#get environment-specifer folder
curate_drive = session.file_by_url("https://drive.google.com/drive/folders/#{drive_id}")

Dir.chdir("#{jobpath}/TO_IIIF")

# Iterate thru the works we are uploading
# if the work is new, create a subfolder for it
# if the work already exists, delete the work xml
# and any other files we are uploading first, then replace them
# with the updates

Dir.glob("#{jobpath}/TO_IIIF/*xml") do |work|
  work_pid = File.basename(work,File.extname(work))
  #Does work dir exist?
  works_folder = nil
  this_work_dir = curate_drive.files( include_team_drive_items: true, q: ["trashed = false AND name = ?", work_pid])

  # if new work, create subfolder else reuse old one,first deleting its contents
  if this_work_dir.length == 1 
    p "#{work_pid}.xml exists"
    works_folder = this_work_dir[0]
    works_folder.files(include_team_drive_items: true, q: "trashed = false") do |file|
      file.delete(permanent: true)
    end
  else
    p "#{work_pid}.xml does not exist"
    works_folder = curate_drive.create_subfolder(work_pid)
  end
  works_folder.upload_from_file("#{work_pid}.xml", "#{work_pid}.xml")
end

