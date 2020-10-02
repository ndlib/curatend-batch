source 'https://rubygems.org'

# This should be everything except :deploy; And by default, we mean any of
# the environments that are not used to execute the deploy scripts
group :default do
  gem 'rake'
  gem 'rof', git: 'https://github.com/ndlib/rof', branch: 'curate-185'
  gem 'resque', '~> 1.23'

  # constrain the following since we want to keep using ruby 2.0
  gem 'rdf-aggregate-repo', '~> 2.0.0'
  gem 'rdf-isomorphic', '~> 2.0.0'
  gem 'rdf-rdfa', '~> 2.0.1'
  gem 'deprecation', '~> 0.2.2'
  gem 'mustermann', '~> 0.3.1'
  gem 'signet', '0.11'
  gem 'google_drive'
  gem 'bundler', '~> 1.17.3'
  
end

group :deploy do
  gem 'capistrano', '~> 2'
end
