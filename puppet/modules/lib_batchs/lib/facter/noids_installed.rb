# Fact: noids_installed
#
# Purpose:
# Return version of noids server installed
#
# Resolution:
# If there is no application user, or cureent symlink, return 'none"
# Else, return whatever /home/app/<app_name>/current sysmlink resolves to
#
 Facter.add('noids_installed') do
	setcode do
 		Facter::Util::Resolution.exec("/opt/noids/bin/noids --version=true |  cut -d' ' -f3")	
	end
 end
