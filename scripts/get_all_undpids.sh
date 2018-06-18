#/bin/bash

# get fedora admin password from FEDORA_PASSWORD
# get fedora_admin user from FEDORA_USER

#Create tempfiles
pidfile=/tmp/pidfile$$
tmpfile=/tmp/tempfile$$

# grab first batch of 100
retcode=$(curl -s -w "%{http_code}"  -o $tmpfile https://${FEDORA_USER}:${FEDORA_PASSWORD}@${fedora_url}/fedora/objects?query=pid%7Eund:*\&maxResults=100\&pid=true\&resultFormat=xml)

#if initial call to fedora fails, return error
if [ ! $retcode -eq '200' ]; then
	echo "Error: Initial fedora call returned $retcode"
	exit 1
fi

# grab successive values until done
while [ $retcode -eq '200' ]; do
    xpath $tmpfile '/result/resultList/objectFields/pid/text()' 2>/dev/null >> $pidfile
    token=$(xpath $tmpfile '/result/listSession/token/text()' 2>/dev/null)
    retcode=$(curl -s -w "%{http_code}"  -o $tmpfile https://${FEDORA_USER}:${FEDORA_PASSWORD}@${fedora_url}/fedora/objects?query=pid%7Eund:*\&maxResults=100\&pid=true\&resultFormat=xml\&sessionToken=${token})
done
 
rm -f $tmpfile

#sanitize pid file
sed -i 0 's/und/ und/g' $pidfile
sed -i 0 's/^ und/und/' $pidfile
cat  $pidfile |  tr ' ' '\n' > /tmp/fedora_allpids

cat /tmp/fedora_allpids
rm $pidfile
rm /tmp/fedora_allpids

exit 0
