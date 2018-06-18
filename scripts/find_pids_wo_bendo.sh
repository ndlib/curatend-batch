#/bin/bash

# get fedora admin password from FEDORA_PASSWORD

for pid in `cat $1`; do
	has_bendo_ds_err=$(curl -s  https://$FEDORA_USER:$FEDORA_PASSWORD@${fedora_url}/fedora/objects/$pid/datastreams/bendo-item?format=xml 2>&1| grep "No datastream could be found"  | wc -l)
	if [ $has_bendo_ds_err -eq "1" ]; then
		echo $pid
	fi
done
