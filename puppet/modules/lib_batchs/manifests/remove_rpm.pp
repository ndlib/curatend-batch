class lib_noids::remove_rpm( $remove_dir = false, $root = 'opt/noids' ) {

		$rpmlist = [ "noids" ]

		# uninstall noids RPM is present
	
		package { $rpmlist:
			ensure => absent,
		}

		# If indicated, remove noids directory tree.

		if $remove_dir == true {

			file {  "${root}":
				ensure => absent,
				recurse => true,
			}
		}
}
