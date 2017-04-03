source 'https://rubygems.org'

# This should be everything except :deploy; And by default, we mean any of
# the environments that are not used to execute the deploy scripts
group :default do
  gem 'rake'
  gem 'rof', github: 'ndlib/rof', branch: 'DLTP-946'
  gem 'resque', '~> 1.23'

  # constrain the following since we want to keep using ruby 2.0
  gem 'rdf-aggregate-repo', '~> 2.0.0'
  gem 'rdf-rdfa', '~> 2.0.1'
end

group :deploy do
  gem 'capistrano', '~> 2'
end
