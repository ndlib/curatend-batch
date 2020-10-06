:quiet# List all tasks from RAILS_ROOT using: cap -T

set :bundle_roles, [:app]
set :bundle_flags, "--deployment"
set :ruby_root, '/opt/rh/rh-ruby26'
require 'bundler/capistrano'

#############################################################
#  Settings
#############################################################

default_run_options[:pty] = true
set :use_sudo, false
ssh_options[:paranoid] = false
set :default_shell, '/bin/bash'

#############################################################
#  SCM
#############################################################

set :scm, :git
set :deploy_via, :remote_cache

#############################################################
#  Environment
#############################################################

namespace :env do
  desc "Set command paths"
  task :set_paths do
    set :bundle_cmd, "#{ruby_root}/root/usr/local/bin/bundle"
    set :rake,      "#{bundle_cmd} exec rake"
  end
end

#############################################################
#  Database
#############################################################

namespace :db do
  desc "Run the seed rake task."
  task :seed, :roles => :app do
    # there is none
  end
end

#############################################################
#  Deploy
#############################################################

namespace :deploy do
  desc "Execute various commands on the remote environment"
  task :debug, :roles => :app do
    run "/usr/bin/env", :pty => false, :shell => '/bin/bash'
    run "whoami"
    run "pwd"
    run "echo $PATH"
    run "which ruby"
    run "ruby --version"
    run "which rake"
    run "rake --version"
    run "which bundle"
    run "bundle --version"
    run "which git"
  end

  desc "Start Application"
  task :start, :roles => :app do
    # done by runsvdir
  end

  desc "Restart application"
  task :restart, :roles => :app do
    # done by runsvdir
  end

  task :stop, :roles => :app do
    # Do nothing.
  end

  desc "Run the migrate rake task."
  task :migrate, :roles => :app do
    # We have no database
  end
end

#############################################################
#  Callbacks
#############################################################

before 'deploy', 'env:set_paths'

#############################################################
#  Configuration
#############################################################

set :application, 'curatend-batch'
set :repository,  "git://github.com/ndlib/curatend-batch.git"

#############################################################
#  Environments
#############################################################

def common
  # can also set :branch with the :tag variable
  set :branch,    fetch(:branch, fetch(:tag, 'master'))
  set :deploy_to, '/home/app/curatend-batch'
  set :user,      'app'
  set :bundle_without, [:development, :test, :debug]

  default_environment['PATH'] = "#{ruby_root}/root/usr/local/bin:/opt/rh/nodejs010/root/usr/bin:$PATH"
  default_environment['LD_LIBRARY_PATH'] = "#{ruby_root}/root/lib64:/opt/rh/nodejs010/root/lib64:/opt/rh/v8314/root/lib64:$LD_LIBRARY_PATH"

  after 'deploy', 'deploy:cleanup'
end

set    :domain,     fetch(:host, 'curate-test1.lc.nd.edu')
server "app@#{domain}", :app
common()

# these tasks are here as placeholders.
# but since the only difference between envrionments is configuration,
# heira will take care of that.
desc "Setup for staging VM"
task :staging do
end

desc "Setup for pre-production deploy"
task :pre_production do
end

desc "Setup for production deploy"
task :production do
end
