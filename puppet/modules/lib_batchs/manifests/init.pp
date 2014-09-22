class lib_batchs( $batchs_root = '/opt/batchs') {

	$pkglist = [ "golang" ]

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

	file { 'logrotate.d/batchs':
		name => '/etc/logrotate.d/batchs',
		replace => true,
		require => File["batchs.conf"],
		content => template('lib_batchs/logrotate.erb'),
	}

	file { "batchs/config.ini":
		name => "${batchs_root}/config.ini",
		replace => true,
		require => File['logrotate.d/batchs'], 
		content => template('lib_batchs/config.ini.erb'),
	}

	exec { "stop-batchs":
		command => "/sbin/initctl stop batchs",
		unless => "/sbin/initctl status batchs | grep stop",
		require => File[ "batchs/config.ini"],
	}

	exec { "start-batchs":
		command => "/sbin/initctl start batchs",
		unless => "/sbin/initctl status batchs | grep process",
		require => Exec["stop-batchs"]
	}
		
}
