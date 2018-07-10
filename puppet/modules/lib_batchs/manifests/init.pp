class lib_batchs( $batchs_root = '/opt/batchs') {

	include lib_runit

	$pkglist = [
		"golang",
		"jq"
	]
	$batch_bendo_url = hiera('batch_bendo_url')
	$batch_jhu_jar = hiera('batch_jhu_jar')
	$batch_osf_host = hiera('batch_osf_host')
	$batch_osf_auth = hiera('batch_osf_auth')
	$batch_queue_dir = hiera('batch_queue_dir')
	$fedora_passwd = hiera('fedora_passwd')
	$batch_curate_url = hiera('batch_curate_url')
	$batch_curate_api_key = hiera('batch_curate_api_key')
	$batch_fedora_url = hiera('batch_fedora_url')
	$batch_noid_pool = hiera('batch_noid_pool')
	$batch_redis_host_port = hiera('batch_redis_host_port')
	$bclient_api_key = hiera('bclient_api_key')
	$solr_corename = hiera('solr_corename')
	$solr_url = hiera('solr_url')

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

	file { 'batchs/scripts':
		name => "${batchs_root}/scripts",
		ensure => 'directory',
	} ->

        # This fetchs the jar every time
	remote_file { "${batchs_root}/scripts/osf-cli.jar":
		remote_location => "${batch_jhu_jar}",
		mode => '0755'
	} ->

	file { 'batchs/scripts/osf-cli-version':
		name => "${batchs_root}/scripts/osf-cli-version",
		replace => true,
		content => template('lib_batchs/cli-version.erb'),
	} ->

	file { 'batchs/tasks/osf-conf':
		name => "${batchs_root}/tasks/osf-cli.conf",
		replace => true,
		content => template('lib_batchs/osf-jhu.conf.erb'),
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

# Start the service and describe it to puppet

	class { 'lib_runit::add':
		service_name => "batchs",
		service_path => "/etc/sv/batchs",
		require => File[['batchsrunitlog', "${batchs_root}/log", "${batchs_root}/tasks/conf" ]],
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
