#!/bin/bash

source /opt/batchs/tasks/conf

#if fedora_password not in environment use what was sourced in
if [[ -z "$FEDORA_USER" || -z "$FEDORA_PASSWORD" ]]; then
	export FEDORA_USER=$fedora_user
	export FEDORA_PASSWORD=$fedora_pass
fi

if [ $# -ne 1 ]; then
	echo "Usage: create_audit_job.sh <directory>"
	exit 1
fi

if [ ! -d $1 ]; then
	echo "Directory $1 does not exist."
	exit 1
fi

echo "create_audit_job  started $(date)"

bindir=$(pwd)

#use audit directory
cd  $1
echo "Using $1"

#if list of all pids exists, use it. If not, create via query of fedora
if [ ! -f ./pids_allund ]; then
	${bindir}/get_all_undpids.sh > ./pids_allund
	if [ $? -ne 0 ]; then
		echo "ERROR: Could Not Retrieve List of Pids From Fedora- aborting"
		exit 1
	fi
fi	

echo "get_all_undpids found $(wc -l ./pids_allund) und: pids"
 
#if list of pids without bendo_item exists, use it. If not, create via query of fedora
if [ ! -f ./pids_nobendo ]; then
	${bindir}/find_pids_wo_bendo.sh ./pids_allund > ./pids_nobendo
	if [ $? -ne 0 ]; then
		echo "ERROR: Could Not Retrieve List of NoBendoItem Pids From Fedora- aborting"
		exit 1
	fi
fi

echo "find_pids_wo_bendo found $(wc -l ./pids_nobendo) pids"
 
#Go through list of Bendo Items , make list of generic files, and of works            
if [ ! -f ./genericfile_pids ]; then
	${bindir}/get-desired-worktypes.sh ./pids_nobendo
	if [ $? -ne 0 ]; then
		echo "ERROR: Could Not compile Lists of NoBendoItem Works and GenericFiles From Fedora- aborting"
		exit 1
	fi
fi

echo "/get-desired-worktypes found $(wc -l ./work_pids) Work and $(wc -l ./genericfile_pids) GenericFiles"

echo "create_audit_job finished $(date)"

exit 0
