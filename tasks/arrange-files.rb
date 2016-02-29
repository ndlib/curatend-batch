#!/usr/bin/env ruby

# read a single rof file (name given on command line)
# and rearrange files for its objects

require 'json'
require 'fileutils'

jobpath = ENV['JOBPATH']
abort("No JOBPATH") if jobpath.nil?

rof_objects = []

rof_file = ARGV[0]
File.open(rof_file, "r") do |f|
  rof_objects.concat(JSON.load(f.read()))
end

bendo_url_re = Regexp.new("^bendo:/item/([^/]+)/(.*)")

rof_objects.each do |obj|
  # if bendo-item entry, save this obj to file system
  bendo_item = obj['bendo-item']
  next unless bendo_item
  pid = obj['pid']
  abort("Missing PID") if pid.nil?
  # remove "und:" prefix if present
  pid = pid.sub(/\Aund:/,"")
  # dump this object into its own rof file
  dst = File.join(jobpath, "TO_TAPE", bendo_item, "fedora3", pid + ".rof")
  puts "Writing #{dst}"
  FileUtils.mkdir_p(File.dirname(dst))
  File.open(dst, "w") do |f|
    # put the object into an array so it matches the rof format
    JSON.dump([obj], f)
  end

  # now scan for URL entries and move files as necessary
  obj.each do |k,v|
    next unless k.end_with?("-meta")
    next unless v['URL']
    # the target bendo item is capture 1
    # the target file path is capture 2
    m = bendo_url_re.match(v['URL'])
    next if m.nil?
    src = File.join(jobpath, m[2])
    dst = File.join(jobpath, "TO_TAPE", m[1], m[2])
    # if the file exists don't overwrite it
    if File.exist?(dst)
      puts "Keeping #{dst}"
      next
    end
    puts "Moving #{dst}"
    FileUtils.mkdir_p(File.dirname(dst))
    FileUtils.mv(src, dst)
  end
end
