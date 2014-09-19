
class lib_curate::worker {

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

     # Install FITS
     class { 'lib_fits':
	require => Package[$packagelist],
     }	


     # Install REDIS
     class { 'lib_redis':
	require => Package[$packagelist],
	config => 'client',
     }	

     # Install app user
     class { 'lib_app_home':
	require => Package[$packagelist],
     }	

     # Install Ruby
     class { 'lib_ruby':
	require => Package[$packagelist],
     }	


     # Install resque worker daemon and service
     class { 'lib_resque_poold':
	require => Class[["lib_ruby","lib_app_home","lib_redis"]],
     }
}
