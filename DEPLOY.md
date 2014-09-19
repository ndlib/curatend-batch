# Deploying the Application

This document describes how the application itself is deployed. It does not
cover the steps necessary to setup the target system. Deploys to Preproduction
and Production are done through Jenkins. Deploys to Staging can be done by any
developer.

# Staging Deploys

You can deploy to any staging machine either through Jenkins or from your command line.
If you want to deploy from the command line, you will need to have given your public ssh key
to Mark Suhovecky to add to the puppet scripts.

### To deploy your working branch to libvirt7

    cap staging deploy -s host=libvirt7.library.nd.edu -s branch=my-working-branch-name

### To deploy the tag `v2013.4` to libvirt7

set either the `branch` or the `tag` variable:

    cap staging deploy -s host=libvirt7.library.nd.edu -s branch=v2013.4

or

    cap staging deploy -s host=libvirt7.library.nd.edu -s tag=v2013.4

### To restart the application on libvirt6

    cap staging deploy:restart -s host=libvirt6.library.nd.edu

### To inspect the logs on libvirt7

    ssh app@libvirt7.library.nd.edu

You will be connected without needing a password (because you gave your
public key to Mark).  The application is deployed to
`~/curatend-batch/current`.  The overlord application is deployed to
`/opt/batch-overlord`.  The logs are in `/opt/batch-overlord/logs/log`.
