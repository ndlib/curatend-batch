source 'https://rubygems.org'

# This should be everything except :deploy; And by default, we mean any of
# the environments that are not used to execute the deploy scripts
group :default do
  gem 'rails'
  gem 'rake'
  gem 'rof', git: 'https://github.com/ndlib/rof'
  gem 'resque', '~> 1.23'
end

group :deploy do
  gem 'capistrano', '~> 2'
end
