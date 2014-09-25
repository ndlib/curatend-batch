class lib_batchs( $batchs_root = '/opt/batchs') {

	$pkglist = [ "golang" ]
	$batch_queue_dir = hiera('batch_queue_dir')
	$fedora_passwd = hiera('fedora_passwd')

	package { $pkglist:
		ensure => present,
	}

	file { [ "$batchs_root", "${batchs_root}/log" ]:
		ensure => directory,
		require => Package["$pkglist"],
	}

	exec { "Build-batchs-from-repo":  
		command => "/bin/bash -c \"export GOPATH=${batchs_root} && go get github.com/ndlib/curatend-batch\"",	
		unless => "/usr/bin/test -f ${batchs_root}/bin/batchs",
		require => File[$batchs_root],
	}

	file { 'batchs.conf':
		name => '/etc/init/batchs.conf',
		replace => true,
		content => template('lib_batchs/upstart.erb'),
		require => Exec["Build-batchs-from-repo"],
	}

	file { 'batchs/tasks':
		name => "${batchs_root}/tasks",
		ensure => 'link',
		target => "${batchs_root}/src/github.com/ndlib/curatend-batch/tasks",
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
