# oxy

`oxy` is a simple program to bootstrap and manage a development environment for the [ALICE-O2](https://github.com/AliceO2Group/AliceO2) project.

## Introduction

`oxy` is a rather opinionated tool to bootstrap and manage a development environment for `O2`.
`oxy` creates a workarea under which you will find:

- `src`: a directory holding the `git` repositories of `FairSoft`, `FairRoot` and `AliceO2`,
- `externals`: a directory holding the binaries+headers for the externals (_ie:_ `FairSoft`, _A.K.A_ `ROOT`, `CMake`, `nanomsg`, ...)
- `sw`: a directory holding the binaries+headers for `FairRoot`

After having initialized and compiled the workarea, the layout of that workarea would look something like:

```sh
$> tree .
.
├── externals
│   ├── bin
│   │   ├── ccmake
│   │   ├── cmake
│   │   ├── cpack
│   │   ├── ctest
│   │   ├── fairsoft-config
│   │   └── ...
│   ├── include
│   │   ├── boost
│   │   ├── FairSoftVersion.h
│   │   ├── gsl
│   │   └── gtest
│   ├── lib
│   │   ├── libboost_atomic.a
│   │   └── ...
│   ├── lib64 -> /opt/alice/sw/externals/lib
│   └── share
│       ├── aclocal
│       ├── cmake-3.3
│       ├── doc
│       ├── info
│       └── ...
├── install
└── src
    ├── alice-o2
    │   ├── README.md
    │   ├── Resources
    │   └── ...
    ├── config.cache
    ├── fair-root
    │   ├── LICENSE
    │   ├── README.md
    │   ├── scripts
    │   ├── templates
    │   ├── test
    │   └── ...
    └── fair-soft
        ├── alfaconfig.sh
        ├── README.md
        ├── scripts
        ├── tools
        └── ...
```

`oxy` uses a `centos:7` container where all the (known and) needed `DEPENDENCIES` to build the whole `AliceO2` software stack are installed.
`oxy build` then proceeds with building the stack from within that container, with the `src` directory of the workarea mounted inside the container under `/opt/alice/src`, `externals` mounted as `/opt/alice/sw/externals` and `install` as `/opt/alice/sw/install`

`oxy shell` allows to spawn commands inside that container (or to spawn an interactive `bash` shell.)

## Example

```sh
$> oxy init work
oxy init work
oxy: init [/home/binet/dev/alice/o2/go/work]...
oxy: fetching repo [git://github.com/FairRootGroup/FairSoft]...
Cloning into '/home/binet/dev/alice/o2/go/work/src/fair-soft'...
remote: Counting objects: 1030, done.
remote: Total 1030 (delta 0), reused 0 (delta 0), pack-reused 1030
Receiving objects: 100% (1030/1030), 17.80 MiB | 7.51 MiB/s, done.
Resolving deltas: 100% (655/655), done.
Checking connectivity... done.
oxy: fetching repo [git://github.com/FairRootGroup/FairRoot]...
Cloning into '/home/binet/dev/alice/o2/go/work/src/fair-root'...
remote: Counting objects: 39765, done.
remote: Compressing objects: 100% (137/137), done.
remote: Total 39765 (delta 89), reused 2 (delta 2), pack-reused 39625
Receiving objects: 100% (39765/39765), 53.53 MiB | 13.94 MiB/s, done.
Resolving deltas: 100% (28681/28681), done.
Checking connectivity... done.
oxy: fetching repo [git://github.com/AliceO2Group/AliceO2]...
Cloning into '/home/binet/dev/alice/o2/go/work/src/alice-o2'...
remote: Counting objects: 34718, done.
remote: Compressing objects: 100% (82/82), done.
remote: Total 34718 (delta 42), reused 0 (delta 0), pack-reused 34633
Receiving objects: 100% (34718/34718), 57.21 MiB | 17.78 MiB/s, done.
Resolving deltas: 100% (26129/26129), done.
Checking connectivity... done.
oxy: config [/home/binet/dev/alice/o2/go/work/src/config.cache]:
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

$> cd work
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
::: sourcing alice-o2 environment...
System during compilation: CentOS Linux release 7.2.1511 (Core) 
                           x86_64
System now               : CentOS Linux release 7.2.1511 (Core) 
                           x86_64
::: sourcing alice-o2 environment... [done]
[oxy@oxy-box alice]$ ex1-sampler --id sampler1 --config-json-file /opt/alice/src/fair-root/examples/MQ/1-sampler-sink/ex1-sampler-sink.json 
[15:32:51][STATE] Entering FairMQ state machine
[15:32:51][INFO] *******************************************************************************************************************************
[15:32:51][INFO] ************************************************     Program options found     ************************************************
[15:32:51][INFO] *******************************************************************************************************************************
[15:32:51][INFO] config-json-file = /opt/alice/src/fair-root/examples/MQ/1-sampler-sink/ex1-sampler-sink.json  [Type=string]  [provided value] *
[15:32:51][INFO] id               = sampler1                                                                   [Type=string]  [provided value] *
[15:32:51][INFO] io-threads       = 1                                                                          [Type=int]     [default value]  *
[15:32:51][INFO] log-color-format = 1                                                                          [Type=bool]    [default value]  *
[15:32:51][INFO] text             = Hello                                                                      [Type=string]  [default value]  *
[15:32:51][INFO] verbose          = DEBUG                                                                      [Type=string]  [default value]  *
[15:32:51][INFO] *******************************************************************************************************************************
[15:32:51][DEBUG] Found device id 'sampler1' in JSON input
[15:32:51][DEBUG] Found device id 'sink1' in JSON input
[15:32:51][DEBUG] [node = device]   id = sampler1
[15:32:51][DEBUG] 	 [node = channel]   name = data-out
[15:32:51][DEBUG] 	 	 [node = socket]   socket index = 1
[15:32:51][DEBUG] 	 	 	 type        = push
[15:32:51][DEBUG] 	 	 	 method      = bind
[15:32:51][DEBUG] 	 	 	 address     = tcp://*:5555
[15:32:51][DEBUG] 	 	 	 sndBufSize  = 1000
[15:32:51][DEBUG] 	 	 	 rcvBufSize  = 1000
[15:32:51][DEBUG] 	 	 	 rateLogging = 0
[15:32:51][DEBUG] ---- Channel-keys found are :
[15:32:51][DEBUG] data-out
[15:32:51][INFO] PID: 35
[15:32:51][INFO] Using nanomsg library
[15:32:51][STATE] Entering INITIALIZING DEVICE state
[15:32:51][INFO] created socket device-commands.pub
[15:32:51][INFO] bind socket device-commands.pub on inproc://commands
[15:32:51][DEBUG] Validating channel "data-out[0]"... VALID
[15:32:51][DEBUG] Initializing channel data-out[0] (push)
[15:32:51][INFO] created socket data-out[0].push
[15:32:51][DEBUG] Binding channel data-out[0] on tcp://*:5555
[15:32:51][INFO] bind socket data-out[0].push on tcp://*:5555
[15:32:51][INFO] created socket device-commands.sub
[15:32:51][INFO] connect socket device-commands.sub to inproc://commands
[15:32:52][STATE] Entering DEVICE READY state
[15:32:52][STATE] Entering INITIALIZING TASK state
[15:32:52][STATE] Entering READY state
[15:32:52][STATE] Entering RUNNING state
[15:32:52][INFO] Use keys to control the state machine:
[15:32:52][INFO] [h] help, [p] pause, [r] run, [s] stop, [t] reset task, [d] reset device, [q] end, [j] init task, [i] init device
[15:32:52][INFO] DEVICE: Running...
[15:32:53][INFO] Sending "Hello"
q
[15:32:56][INFO] [q] end
[15:32:56][STATE] Entering EXITING state
[15:32:56][DEBUG] unblocked
[15:32:56][DEBUG] Closing sockets...
[15:32:56][DEBUG] Closed all sockets!
[15:32:56][STATE] Exiting FairMQ state machine
[oxy@oxy-box alice]$ exit
$>
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
