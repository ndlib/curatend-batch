#!/bin/bash
# search target batch ingest environment for reingested ETDs 
# Make a pids/bendo id mapping of them, and reingest yet again
# Should be run as app user on worker VM (pprd or prod) 

if [ $# -ne 1 ]
then
        echo "Usage: find_etd_pids.sh <env>"
        echo "       env is prod | pprd | libvirt[689]"
        exit 1
fi

if [ $UID -ne 1518 ]
then
        echo "Command must be run as app user"
        exit 1
fi

# check command line args
case  $1 in
prod)
	BATCH_INGEST_ROOT="/mnt/curatend-batch/production"
	;;
pprd|libvirt[689])
	BATCH_INGEST_ROOT="/mnt/curatend-batch/test/$1"
	;;
*)
        echo "Usage: find_etd_pids.sh <env>"
        echo "       env is prod | pprd | libvirt[689]"
        exit 1
esac

DATESTAMP=$(date +%Y%h%d%H%M%S)

# search for reingested ETDs
grep -R -l ms:degree $BATCH_INGEST_ROOT/success/reingest*/FROM_TAPE > /tmp/etd.out${DATESTAMP}

NUM_ETD=$(cat /tmp/etd.out${DATESTAMP} | wc -l)

#exit if nobe were found
if [ $NUM_ETD -eq 0 ]
then
	echo "No ETDS Found"
	exit 0
fi


#make staging directory
mkdir -p $BATCH_INGEST_ROOT/data/etd_remediate-${DATESTAMP}

#generate fedora-pids mapping with bendo items
cat /tmp/etd.out${DATESTAMP} | sed 's/.rof$//' | awk 'BEGIN{FS="/"; printf "{"}{ printf "\"und:%s\":\"%s\",",$10,$8}END{printf "}"}' | sed 's/,}$/}/' > $BATCH_INGEST_ROOT/data/etd_remediate-${DATESTAMP}/fedora-pids

#generate JOBS file
cat <<EOF  > $BATCH_INGEST_ROOT/data/etd_remediate-${DATESTAMP}/JOB
{"Todo":["fedora-to-rof",
 "get-from-bendo",
 "compare-rof",
 "validate",
 "move-files-for-bendo",
 "upload-to-bendo",
 "ingest",
 "index"]}
EOF

#queue up job
mv  $BATCH_INGEST_ROOT/data/etd_remediate-${DATESTAMP}  $BATCH_INGEST_ROOT/queue

rm /tmp/etd.out${DATESTAMP}
