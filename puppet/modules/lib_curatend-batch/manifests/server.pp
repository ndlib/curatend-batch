
class lib_curate::server {

   # lookup data we need to build this node
   $ssl_certificate = hiera('ssl_certificate')
   $ssl_certificate_key = hiera('ssl_certificate_key')
   $env = hiera('env')


   # Set up the Common Pacakges needed for CurateND
   # Binary package requirements
	
	$packagelist = [
    "clamav",
    "clamav-db",
    "clamav-devel",
    "ImageMagick",
    "GConf2",
    "ORBit2",
    "libIDL",
    "libcroco",
    "libgcrypt-devel",
    "libgpg-error-devel",
    "libgsf",
    "librsvg2",
    "libtool-ltdl",
    "libwmf-lite",
    "libxml2-devel",
    "libxslt-devel",
    "readline-devel",
    "noids",
    "mysql-devel",
    "sgml-common",
    "htop",
    "tmux" ]

	package { $packagelist:
        	ensure => installed
	}


     # Install REDIS (client and server)
     class { 'lib_redis':
	require => Package[$packagelist],
	config => "server",
	type => "remote",
     }	

     # Install SSL Certs
     class { 'lib_certs':
	require => Package[$packagelist],
     }	

     # Install app user
     class { 'lib_app_home':
	require => Package[$packagelist],
     }	

     # Install Ruby
     class { 'lib_ruby':
	require => Package[$packagelist],
     }	

     class { 'lib_nginx':
	require => Class[["lib_app_home","lib_certs","lib_ruby"]],
     }

     file { '/etc/nginx/conf.d/curatend.conf':
       content => template('lib_curate/curatend.conf.erb'),
       owner => 'root',
       group => 'root',
       mode => '644',
       notify => Service['nginx'],
     }

     file { '/etc/nginx/conf.d/default.conf':
       ensure => absent,
       require => File['/etc/nginx/conf.d/curatend.conf'],
       notify => Service['nginx'],
     }

     # Config file for noids server
     file { '/opt/noids/config.ini':
	ensure => present,
	replace => true,
	source => "puppet:///modules/lib_curate/config.ini.${env}",
	require => Package["noids"],
      }


}
