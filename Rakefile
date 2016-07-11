# Add your own tasks in files placed in lib/tasks ending in .rake,
# for example lib/tasks/capistrano.rake, and they will automatically be available to Rake.

require File.expand_path('../config/application', __FILE__)

Rails.application.load_tasks

if defined?(RSpec)
  namespace :spec do
    desc "Run all specs"
    RSpec::Core::RakeTask.new(:all) do
      ENV['COVERAGE'] = 'true'
    end

    namespace :coverage do
      desc "Run all non-feature specs"
      RSpec::Core::RakeTask.new(:without_features) do |t|
        ENV['COVERAGE'] = 'true'
        t.exclude_pattern = './spec/features/**/*_spec.rb'
        t.rspec_opts = '--profile 10'
      end

      types = begin
        dirs = Dir['./app/**/*.rb'].map { |f| f.sub(%r{^\./(app/\w+)/.*}, '\\1') }.uniq.select { |f| File.directory?(f) }
        Hash[dirs.map { |d| [d.split('/').last, d] }]
      end

      types.each do |name, _dir|
        desc "Run, with code coverage, the examples in spec/#{name.downcase}"
        RSpec::Core::RakeTask.new(name) do |t|
          ENV['COVERAGE'] = 'true'
          ENV['COV_PROFILE'] = name.downcase
          t.pattern = "./spec/#{name}/**/*_spec.rb"
        end
      end
    end

    desc 'Run the Travis CI specs'
    task :travis do
      ENV['SPEC_OPTS'] ||= "--profile 5"
      Rake::Task[:default].invoke
    end
  end
end

# BEGIN `commitment:install` generator
# This was added via commitment:install generator. You are free to change this.
Rake::Task["default"].clear
task(
  default: [
    'commitment:rubocop',
    'commitment:configure_test_for_code_coverage',
    'spec',
    'commitment:code_coverage',
    'commitment:brakeman'
  ]
)
# END `commitment:install` generator
