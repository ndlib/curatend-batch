#
# This class is called from Capistrano, and has the overall layout
# of the batch server

class lib_curatend_batch() {

	# create app user, build ruby

	include lib_app_home
	include lib_ruby
	include lib_batchs

	# app subdirectory for batch

	file { "/home/app/curatend-batch": 
		ensure => directory,
		mode => 0755,
		owner => "app",
		group => "app",
		require => Class["lib_app_home"],
	}

}