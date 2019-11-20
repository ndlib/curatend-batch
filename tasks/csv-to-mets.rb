#!/usr/bin/env ruby
# frozen_string_literal: true

require 'csv'
require 'erb'

# Utility Class to properly bind the CSV data er read in
# to pass to the ERB template
#
class IiifData
  attr_accessor :work_data, :sequence_data, :work_dir, :iiif_env

  def initialize(main_table, sequence_table, workdir, iiifenv)
    @sequence_data = sequence_table
    @work_data = main_table
    @work_dir = workdir
    @iiif_env = iiifenv
  end

  # this is what ERB::result calls to get data
  def get_binding
    binding()
  end
end

# JOBPATH and IIIF_ENV  will be set by batch ingest.
jobpath = ENV['JOBPATH']
iiif_env = ENV['IIIF_ENV']

# if $JOBPATH/TO_IIIF dir does not exist, we are done, and can exit

unless File.directory?("#{jobpath}/TO_IIIF")
  p 'No works to upload to IIIF pipeline.'
  exit 0
end

Dir.chdir("#{jobpath}/TO_IIIF")

# Iterate through subdirectories under TO_IIIF-one per work, named by CurateND pid

Dir.glob('*') do |work_dir|
  main_table = CSV.parse(File.read("#{jobpath}/TO_IIIF/#{work_dir}/main.csv"), {:quote_char => "\'", :headers => true})
  sequence_table = CSV.parse(File.read("#{jobpath}/TO_IIIF/#{work_dir}/sequence.csv"), {:quote_char => "\'", :headers => true})

  # instantiate iiif binding class to pass data to template
  iiif_data = IiifData.new(main_table, sequence_table, work_dir, iiif_env)

  # read in and build template
  mets_xml = ERB.new(File.read("/opt/batchs/tasks/iiif_template.erb"))
  mets_xml_output = mets_xml.result(iiif_data.get_binding)

  # Write it to proper directory
  File.write("#{jobpath}/TO_IIIF/#{work_dir}/#{work_dir}.xml", mets_xml_output)
end
