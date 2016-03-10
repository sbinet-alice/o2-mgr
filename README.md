# oxy

`oxy` is a simple program to bootstrap and manage a development environment for the [ALICE-O2](https://github.com/AliceO2Group/AliceO2) project.

## Example

```sh
$> oxy init dev
oxy init dev
oxy: init [/home/binet/dev/alice/o2/go/src/github.com/sbinet-alice/oxy/dev]...
oxy: fetching repo [git://github.com/FairRootGroup/FairSoft]...
Cloning into '/home/binet/dev/alice/o2/go/src/github.com/sbinet-alice/oxy/dev/src/fair-soft'...
remote: Counting objects: 1030, done.
remote: Total 1030 (delta 0), reused 0 (delta 0), pack-reused 1030
Receiving objects: 100% (1030/1030), 17.80 MiB | 7.51 MiB/s, done.
Resolving deltas: 100% (655/655), done.
Checking connectivity... done.
oxy: fetching repo [git://github.com/FairRootGroup/FairRoot]...
Cloning into '/home/binet/dev/alice/o2/go/src/github.com/sbinet-alice/oxy/dev/src/fair-root'...
remote: Counting objects: 39765, done.
remote: Compressing objects: 100% (137/137), done.
remote: Total 39765 (delta 89), reused 2 (delta 2), pack-reused 39625
Receiving objects: 100% (39765/39765), 53.53 MiB | 13.94 MiB/s, done.
Resolving deltas: 100% (28681/28681), done.
Checking connectivity... done.
oxy: fetching repo [git://github.com/AliceO2Group/AliceO2]...
Cloning into '/home/binet/dev/alice/o2/go/src/github.com/sbinet-alice/oxy/dev/src/alice-o2'...
remote: Counting objects: 34718, done.
remote: Compressing objects: 100% (82/82), done.
remote: Total 34718 (delta 42), reused 0 (delta 0), pack-reused 34633
Receiving objects: 100% (34718/34718), 57.21 MiB | 17.78 MiB/s, done.
Resolving deltas: 100% (26129/26129), done.
Checking connectivity... done.
oxy: config [/home/binet/dev/alice/o2/go/src/github.com/sbinet-alice/oxy/dev/src/config.cache]:
## config for FairRoot externals
compiler=gcc
debug=no
optimize=yes
geant4_download_install_data_automatic=no
geant4_install_data_from_dir=no
build_root6=yes
build_python=yes
install_sim=yes
SIMPATH_INSTALL=/opt/alice/sw/externals
platform=linux
oxy: init-container [oxy-dev]...
oxy: retrieving latest centos:7 image...
7: Pulling from library/centos
a3ed95caeb02: Already exists 
a3ed95caeb02: Already exists 
Digest: sha256:3cdc0670fe9130ab3741b126cfac6d7720492dd2c1c8ae033dcd77d32855bab2
Status: Image is up to date for centos:7
oxy: building container [oxy-dev]...
Sending build context to Docker daemon  2.56 kB
Step 1 : FROM centos:7
 ---> d0e7f81ca65c
Step 2 : RUN yum update -y && yum install -y cmake gcc gcc-c++ gcc-gfortran make patch sed libX11-devel libXft-devel libXpm-devel libXext-devel libXmu-devel mesa-libGLU-devel mesa-libGL-devel ncurses-devel curl bzip2 libbz2-dev gzip unzip tar expat-devel subversion git flex bison imake redhat-lsb-core python-devel libxml2-devel wget openssl-devel curl-devel automake autoconf libtool which
 ---> Using cache
 ---> efa61490398a
Step 3 : RUN useradd -m -u 1000 oxy-dev
 ---> Using cache
 ---> b9db2e9a78d7
Step 4 : USER oxy-dev
 ---> Using cache
 ---> 43e36c2e8c0f
Successfully built 43e36c2e8c0f

$> cd dev
$> oxy build
oxy: building fair-soft...
The build process for the external packages for the FairRoot Project was started at 100316_120559
[...]

oxy: building fair-root...
[...]

oxy: building alice-o2...
[...]

oxy: ::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
oxy: oxy build complete
oxy: run:
$> source /opt/alice/src/alice-o2/build/config.sh

oxy: to get a runtime environment.
oxy: build complete.
oxy: bye.

$> oxy shell

```

## Documentation

```sh
$> oxy help
oxy - manages a development environment

Commands:

    build          build target(s)
    init           initialize a new workarea
    init-container init a container
    shell          run a shell

Use "oxy help <command>" for more information about a command.
```
