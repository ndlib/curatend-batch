# Contributing to CurateND-Batch

## Getting Started
CurateND-batch is a go service. There are several steps you must complete before you can to run the services, compile the code, and commit to the project.

### Development Dependencies
CurateND-Batch depends on `go` and `jq` in order to work correctly.

```console
brew update
brew install go jq
```

> These homebrew recipes are already included [DLT dotfiles](https://github.com/ndlib/dlt-dotfiles).

### Preparing the Environment
Go has built-in dependency management. It performs functions inside a directory that is set as the `$GOPATH`. Our convention for development environments is to set `$GOPATH` to `~/gocode`.

```console
mkdir ~/gocode
export GOPATH='~/gocode'
```

> If you manage your shell environment with [DLT dotfiles](https://github.com/ndlib/dlt-dotfiles) it will set up your `$GOPATH` for you.

To verify that `$GOPATH` is configured correctly try:

```console
echo $GOPATH
```

It should return `/Users/<YOUR_USERNAME>/gocode`.

### Checking out the Codebase
Once `$GOPATH` is configured use `go get` to check out the git repository and keep track of it so it can be included by other go projects.

```console
go get -u github.com/ndlib/curatend-batch
```

### Configuring git
The default configuration for git repositories created via `go get` is not set up to allow you to make commits back to the project. Before you make changes to the CurateND-batch codebase you will need to reconfigure the git repository.

```console
cd $GOPATH/src/github.com/ndlib/curatend-batch
git remote set-url origin git@github.com:ndlib/curatend-batch.git
```

> This remote URL assumes you have commit access to ndlib/curatend-batch. If you are working on another repo, like a fork, use the URL provided buy github for that repo.

### Building CurateND-Batch
CurateND-Batch defines one executable `curatend-batch`. When the codebase in checked out using `go get` it is compiled and placed in `$GOPATH/bin`.

> `$GOPATH/bin` is already included in your `$PATH` if you use [DLT dotfiles](https://github.com/ndlib/dlt-dotfiles). Otherwise you will want to call it directly `$GOPATH/bin/curatend-batch` or include `$GOPATH/bin` in you `$PATH` manually e.g. `export $PATH=$GOPATH/bin:$PATH`.

#### If you have made changes locally
To recompile the code after you make changes use `go build`. It will create the `curatend-batch` executable at the root of the project directory. You will have to manually update the executable in `$GOPATH/bin`.

```console
cd $GOPATH/src/github.com/ndlib/curatend-batch/
go build
mv curatend-batch $GOPATH/bin/
```

#### If you want to install the latest version from github
If you have already pushed your changes to github or if the project has been updated and you want to install the latest version use `go get`.

```console
go get -u github.com/ndlib/curatend-batch
```

### Installing the ROF gem
Much of the logic performed in CurateND-batch is housed in the [ROF gem](https://github.com/ndlib/rof). It will need to be installed.

```console
cd $GOPATH/src/github.com/ndlib/curatend-batch/ && bundle
```

> This assumes you already have ruby, rubygems, and bundler installed. [DLT-dotfiles](https://github.com/ndlib/dlt-dotfiles/) takes care of this for you—otherwise it is an exercise left to the reader.

## Running CurateND-Batch
There are also several setup steps in order to _run_ the batch ingest.

### Directory Setup
By convention we will run CurateND-Batch out of the `~/goapps` directory. It expects the presence of to directories: `queue` and `tasks`.

```console
mkdir -p ~/goapps/curatend-batch/development
```

### Starting the Application
There are two parts of the CurateND-Batch application: a job manager and a task runner.

#### Starting the Job Manager
The job manager is go service, packaged as the `curatend-batch` executable. It can be run from any directory. By default it will log to STDOUT. There is a bootstrap script that will create the appropriate directories and start the services.

```console
cd $GOPATH/src/github.com/ndlib/curatend-batch/ && ./bootstrap.sh
```

#### Starting the Task Runner
The task runner is a combination of shell scripts, ruby scrips, and logic contained within the [ROF gem](https://github.com/ndlib/rof). It must be run from the `tasks` directory (the same one passed to `curatend-batch`). When it is initialized it retrieves the configuration information from `tasks/config`. It has a bunch of dependencies:

- [CurateND](https://github.com/ndlib/curate_nd)
- Fedora (use the [hydra-jetty](https://github.com/projecthydra/hydra-jetty) in CurateND)
- [curatend-batch](https://github.com/ndlib/curatend-batch)
- [noids](https://github.com/ndlib/noids)
- Redis (use homebrew: `brew install redis`; [DLT-dotfiles](https://github.com/ndlib/dlt-dotfiles/) does this for you)

Once you have all of this running locally you can start running batch ingest jobs.

> NOTE: Jobs are run in subshells that have very little configuration information. This means that jobs are unaware of ruby managers like [rvm](https://rvm.io/), [rbenv](https://github.com/rbenv/rbenv), and [chruby](https://github.com/postmodern/chruby). One way to get around this is to install the project dependencies in the “system” ruby. (This is not a particularly good solution.)