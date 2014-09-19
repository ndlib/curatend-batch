:quiet# List all tasks from RAILS_ROOT using: cap -T
#
# NOTE: The SCM command expects to be at the same path on both the local and
# remote machines. The default git path is: '/shared/git/bin/git'.

set :bundle_roles, [:app, :work]
set :bundle_flags, "--deployment"
require 'bundler/capistrano'
# see http://gembundler.com/v1.3/deploying.html
# copied from https://github.com/carlhuda/bundler/blob/master/lib/bundler/deployment.rb
#
# Install the current Bundler environment. By default, gems will be \
#  installed to the shared/bundle path. Gems in the development and \
#  test group will not be installed. The install command is executed \
#  with the --deployment and --quiet flags. If the bundle cmd cannot \
#  be found then you can override the bundle_cmd variable to specifiy \
#  which one it should use. The base path to the app is fetched from \
#  the :latest_release variable. Set it for custom deploy layouts.
#
#  You can override any of these defaults by setting the variables shown below.
#
#  N.B. bundle_roles must be defined before you require 'bundler/#{context_name}' \
#  in your deploy.rb file.
#
#    set :bundle_gemfile,  "Gemfile"
#    set :bundle_dir,      File.join(fetch(:shared_path), 'bundle')
#    set :bundle_flags,    "--deployment --quiet"
#    set :bundle_without,  [:development, :test]
#    set :bundle_cmd,      "bundle" # e.g. "/opt/ruby/bin/bundle"
#    set :bundle_roles,    #{role_default} # e.g. [:app, :batch]
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
    set :bundle_cmd, '/opt/ruby/current/bin/bundle'
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
    # done by upstart
  end

  desc "Restart application"
  task :restart, :roles => :app do
    # done by upstart
  end

  task :stop, :roles => :app do
    # Do nothing.
  end

  desc "Run the migrate rake task."
  task :migrate, :roles => :app do
    # We have no database
  end
end


namespace :und do
  def run_puppet(options={})
    local_module_path = File.join(release_path, 'puppet', 'modules')
    option_string = options.map { |k,v| "#{k} => '#{v}'" }.join(', ')
    run %Q{sudo puppet apply --modulepath=#{local_module_path}:/global/puppet_standalone/modules:/etc/puppet/modules -e "class { 'lib_curatend-batch': #{option_string} }"}
  end

  desc "Run puppet using the modules supplied by the application"
  task :puppet, :roles => [:app, :work] do
    run_puppet()
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

  default_environment['PATH'] = '/opt/ruby/current/bin:$PATH'

  before 'bundle:install', 'und:puppet'
  after 'deploy', 'deploy:cleanup'
end

set    :domain,     fetch(:host, 'libvirt6.library.nd.edu')
server "app@#{domain}", :app, :work, :web, :db, :primary => true
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
