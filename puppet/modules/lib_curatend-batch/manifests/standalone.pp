class lib_curate::standalone {

   # lookup data we need to build this node
   $env = hiera('env')
   $fedora_admin_mysql = hiera('fedora_admin_mysql')
   $fedora_db_name = hiera('fedora_db_name')
   $fedora_passwd = hiera('fedora_passwd')
   $fedora_passwd_mysql = hiera('fedora_passwd_mysql')
   $fedora_root = hiera('fedora_root')
   $fedora_version = hiera('fedora_version')
   $mysql_db_name = hiera('mysql_db_name')
   $mysql_root_password = hiera('mysql_root_password')
   $ssl_certificate = hiera('ssl_certificate')
   $ssl_certificate_key = hiera('ssl_certificate_key')
   $solr_conf_dir = hiera('solr_conf_dir')
   $solr_corename = hiera('solr_corename')
   $solr_data_dir = hiera('solr_data_dir')
   $solr_instance_dir = hiera('solr_instance_dir')
   $solr_root = hiera('solr_root')
   $solr_version = hiera('solr_version')
   $tomcat_config_dir = hiera('tomcat_config_dir')
   $tomcat_root = hiera('tomcat_root')

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
    "mysql-devel",
    "sgml-common",
    "htop",
    "tmux" ]

	package { $packagelist:
        	ensure => installed
	}

     package { 'noids':
	ensure => absent,
     }

     # Install FITS
     class { 'lib_fits':
	require => Package[$packagelist],
     }	


     # Install REDIS
     class { 'lib_redis':
	require => Package[$packagelist],
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


     # Install and Configure mysql for fedora

     exec { "install-puppet-module":
	unless => "/usr/bin/stat /etc/puppet/modules/mysql",
	command => "/usr/bin/puppet module install puppetlabs/mysql",
	logoutput => on_failure,
	user => "root",
	require => Package[$packagelist],
     } ->
     class { 'mysql::server':
  		root_password =>  $mysql_root_password,
		require => [Package[$packagelist], Class["lib_ruby"], Class["lib_app_home"]],
     } ->
     mysql::db { "${fedora_db_name}":
	user => $fedora_admin_mysql,
	password => $fedora_passwd_mysql,
	host => 'localhost',
	grant => ['all'],
	require => Class['mysql::server'],
     } ->
     mysql::db { 'curate_staging':
	user => 'rails',
	password => 'rails',
	host => 'localhost',
	grant => ['all'],
	require => Class['mysql::server'],
     }

     #Install and start Tomcat6
     class { 'lib_tomcat6':
	require => Package[$packagelist],
     }	

     # Install Fedora after Tomcat and mysql set up
     class { 'lib_fedora':
	require => [Class["lib_tomcat6"], Mysql::Db["${fedora_db_name}"]],
     }

     # Install SOLR after Fedora- tell tomcat
     class { 'lib_solr':
	require => Class["lib_fedora"],
     }

     class { 'lib_nginx':
	require => Class["lib_solr"],
     }

     file { '/etc/nginx/conf.d/curatend.conf':
       content => template('lib_curate/curatend.conf.staging.erb'),
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

     # Install resque worker daemon and service
     class { 'lib_resque_poold':
	require => Class["lib_ruby"],
     }

     # Make unicorn a service ( unicorn.rb is deployed with the ruby stack
     class { 'lib_unicornd':
	require => Class["lib_resque_poold"],
     }

     # Config file for noids server
     file { '/opt/noids/config.ini':
	ensure => present,
	replace => true,
	source => "puppet:///modules/lib_curate/config.ini.${env}",
	require => Package["noids"],
      }

     # set logrotate for application and unicorn log files
     class { 'lib_logrotate::app_log': postrotate => "pkill -USR1 -u app -f 'unicorn master' || true" }
}
