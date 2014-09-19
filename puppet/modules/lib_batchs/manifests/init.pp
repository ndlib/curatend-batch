class lib_noids( $noid_root = '/opt/noids') {

        $pkglist = [ "golang" ]

	class { 'lib_noids::remove_rpm':
		root => $noid_root,
	}

	package { $pkglist:
		ensure => present,
		require => Class["lib_noids::remove_rpm"],
	}

        file { [ "$noid_root", "${noid_root}/log" ]:
		ensure => directory,
		require => Package["$pkglist"],
	}

	exec { "Build-noids-from-repo":  
		command => "/bin/bash -c \"export GOPATH=${noid_root} && go get github.com/ndlib/noids\"",	
		unless => "/usr/bin/test -f ${noid_root}/bin/noids",
		require => File[$noid_root],
        }

	file { 'noids.conf':
		name => '/etc/init/noids.conf',
		replace => true,
		content => template('lib_noids/noids.conf.erb'),
		require => Exec["Build-noids-from-repo"],
	}

	file { 'logrotate.d/noids':
		name => '/etc/logrotate.d/noids',
		replace => true,
		require => File["noids.conf"],
		content => template('lib_noids/noids.erb'),
	}

	file { "noids/config.ini":
		name => "${noid_root}/config.ini",
		replace => true,
		require => File['logrotate.d/noids'], 
		content => template('lib_noids/config.ini.erb'),
	}

        exec { "stop-noids":
		command => "/sbin/initctl stop noids",
		unless => "/sbin/initctl status noids | grep stop",
		require => File[ "noids/config.ini"],
	}

        exec { "start-noids":
		command => "/sbin/initctl start noids",
		unless => "/sbin/initctl status noids | grep process",
		require => Exec["stop-noids"]
	}
		
}
