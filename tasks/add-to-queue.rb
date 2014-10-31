#!/usr/bin/env ruby

require('base64')
require('resque')

# This adds a characterize job to the queue for every id on stdin
# (ids are separated by newlines)

class CharacterizeJob
  def initialize(pid)
    @pid = pid
  end

  def queue_name
    :characterize
  end
end

module Sufia
  module Resque
    class Queue
      def push(job)
        queue = job.queue_name
        ::Resque.enqueue_to queue, MarshaledJob, Base64.encode64(Marshal.dump(job))
      end
    end

    class MarshaledJob
    end
  end
end

redis_host_port = ENV['REDIS_HOST_PORT']
if redis_host_port.nil?
  abort("REDIS_HOST_PORT unset")
end

Resque.redis = redis_host_port
queue = Sufia::Resque::Queue.new

STDIN.each do |line|
  queue.push(CharacterizeJob.new(line.strip))
end

