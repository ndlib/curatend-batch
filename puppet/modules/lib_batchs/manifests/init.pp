class lib_batchs( $batchs_root = '/opt/batchs') {

	$pkglist = [
		"golang",
		"jq"
	]
	$batch_queue_dir = hiera('batch_queue_dir')
	$fedora_passwd = hiera('fedora_passwd')
	$batch_curate_url = hiera('batch_curate_url')
	$batch_curate_api_key = hiera('batch_curate_api_key')
	$batch_fedora_url = hiera('batch_fedora_url')
	$batch_noid_pool = hiera('batch_noid_pool')
	$batch_redis_host_port = hiera('batch_redis_host_port')

	package { $pkglist:
		ensure => present,
	}

	file { [ "$batchs_root", "${batchs_root}/log" ]:
		ensure => directory,
		require => Package[$pkglist],
	}

    # symlink to whatever the jenkis capistrano deploy checked out
    # (to support the deploy tag used by jenkins)
    file { 'batch_go_source':
         path => "${batchs_root}/src/github.com/ndlib/curatend-batch",
         ensure => link,
         target => "/home/app/curatend-batch/current",
         force => true,
         require => File["${batchs_root}"],
    }

	exec { "Build-batchs-from-repo":
		command => "/bin/bash -c \"export GOPATH=${batchs_root} && go get -u github.com/ndlib/curatend-batch\"",
		require => File['batch_go_source'],
	}

	file { 'batchs.conf':
		name => '/etc/init/batchs.conf',
		replace => true,
		content => template('lib_batchs/upstart.erb'),
		require => Exec["Build-batchs-from-repo"],
	}

	file { 'batchs/tasks':
		name => "${batchs_root}/tasks",
		ensure => 'directory',
		source => "${batchs_root}/src/github.com/ndlib/curatend-batch/tasks",
		recurse => true,
		purge => true,
		require => Exec['Build-batchs-from-repo'],
	}

	file { 'batchs/tasks/conf':
		name => "${batchs_root}/tasks/conf",
		replace => true,
		content => template('lib_batchs/tasks.conf.erb'),
		require => File['batchs/tasks'],
	}

	file { 'logrotate.d/batchs':
		name => '/etc/logrotate.d/batchs',
		replace => true,
		require => File["batchs/tasks/conf"],
		content => template('lib_batchs/logrotate.erb'),
	}

	exec { "stop-batchs":
		command => "/sbin/initctl stop batchs",
		unless => "/sbin/initctl status batchs | grep stop",
		require => File['logrotate.d/batchs'],
	}

	exec { "start-batchs":
		command => "/sbin/initctl start batchs",
		unless => "/sbin/initctl status batchs | grep process",
		require => Exec["stop-batchs"]
	}

}
