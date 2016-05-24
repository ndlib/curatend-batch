class lib_batchs( $batchs_root = '/opt/batchs') {

	include lib_runit

	$pkglist = [
		"golang",
		"jq"
	]
	$batch_bendo_url = hiera('batch_bendo_url')
	$batch_queue_dir = hiera('batch_queue_dir')
	$fedora_passwd = hiera('fedora_passwd')
	$batch_curate_url = hiera('batch_curate_url')
	$batch_curate_api_key = hiera('batch_curate_api_key')
	$batch_fedora_url = hiera('batch_fedora_url')
	$batch_noid_pool = hiera('batch_noid_pool')
	$batch_redis_host_port = hiera('batch_redis_host_port')
	$bclient_api_key = hiera('bclient_api_key')

	package { $pkglist:
		ensure => present,
	} ->

	# install bendo bclient into batchs_root
	class { 'lib_go::build':
		repo => 'github.com/ndlib/bendo',
		goroot => "${batchs_root}",
		target => "github.com/ndlib/bendo/cmd/bclient",
	} ->

	file { "${batchs_root}/log":
		ensure => directory,
		owner => "app",
		group => "app",
	} ->

	file { "${batchs_root}/src/github.com/ndlib":
		ensure => directory,
	} ->

	# symlink to whatever the jenkis capistrano deploy checked out
	# (to support the deploy tag used by jenkins)
	file { 'batch_go_source':
		path => "${batchs_root}/src/github.com/ndlib/curatend-batch",
		ensure => link,
		target => "/home/app/curatend-batch/current",
		force => true,
	} ->

	exec { "Build-batchs-from-repo":
		command => "/bin/bash -c \"export GOPATH=${batchs_root} && go get -d github.com/ndlib/curatend-batch && go install github.com/ndlib/curatend-batch\"",
	} ->

	file { 'batchs/tasks':
		name => "${batchs_root}/tasks",
		ensure => 'directory',
		source => "${batchs_root}/src/github.com/ndlib/curatend-batch/tasks",
		recurse => true,
		purge => true,
	} ->

	file { 'batchs/tasks/conf':
		name => "${batchs_root}/tasks/conf",
		replace => true,
		content => template('lib_batchs/tasks.conf.erb'),
	}

# Create batchs runit service directories

	$batchrunitdirs = [ "/etc/sv/batchs", "/etc/sv/batchs/log" ]

	file { $batchrunitdirs:
		ensure => directory,
		owner => "app",
		group => "app",
		require => Class[['lib_runit','lib_go::build']],
	} ->

	file { 'batchsrunitexec':
		name => '/etc/sv/batchs/run',
		owner => "app",
		group => "app",
		mode => '0755',
		replace => true,
		content => template('lib_batchs/run.erb'),
		require => File[$batchrunitdirs],
	} ->

	file { 'batchsrunitlog':
		name => '/etc/sv/batchs/log/run',
		owner => "app",
		group => "app",
		replace => true,
		mode => '0755',
		content => template('lib_batchs/log_run.erb'),
	}

# make sure the upstart managed service is down and service file removed
# (these puppets steps will be unnecessary once batchs has been
# deployed to all the environments)

	exec { 'stop-batchs-upstart':
		command => "/sbin/initctl stop batchs",
		returns => [0, 1],
		require => File[$batchrunitdirs],
	} ->
	file { '/etc/init/batchs':
		ensure => absent,
	}

# Start the service and describe it to puppet

	class { 'lib_runit::add':
		service_name => "batchs",
		service_path => "/etc/sv/batchs",
		require => File[['batchsrunitlog', "${batchs_root}/log", "${batchs_root}/tasks/conf", '/etc/init/batchs']],
	} ->

	service { 'batchs':
		provider => 'base',
		ensure => running,
		enable => true,
		hasstatus => false,
		hasrestart => false,
		restart => '/sbin/sv restart batchs',
		start => '/sbin/sv start batchs',
		stop => '/sbin/sv stop batchs',
		status => '/sbin/sv status batchs',
	}
}
