#/bin/bash

# get fedora admin password from FEDORA_PASSWORD

for pid in `cat $1`; do
	work_type=$( curl -s  https://$FEDORA_USER:$FEDORA_PASSWORD@$fedora_url}/fedora/objects/$pid/datastreams/RELS-EXT/content?format=xml | grep hasModel | grep afmodel | cut -d\" -f 2 | sed "s/^.*rdf:resource='//" | sed "s/'.*>$//" )

	# Make lists of Works and Files not in bendo- discard otherwise
	case  $work_type in

	  "info:fedora/afmodel:GenericFile" ) echo $pid >> genericfile_pids 
		;;
	  "info:fedora/afmodel:Article" | "info:fedora/afmodel:Dataset" | "info:fedora/afmodel:Document" ) echo $pid >> work_pids 
		;;
	  "info:fedora/afmodel:Etd" | "info:fedora/afmodel:FindingAid" | "info:fedora/afmodel:Image" ) echo $pid >> work_pids 
		;;
	  "info:fedora/afmodel:Patent" | "info:fedora/afmodel:SeniorThesis" | "info:fedora/afmodel:Video" ) echo $pid >> work_pids 
		;;
	  "info:fedora/afmodel:LibraryCollection" ) echo $pid >> work_pids 
		;;
	  *)  echo "$pid $work_type"  >> other_pids
		;;
	esac
done
