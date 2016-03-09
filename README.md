# o2-mgr

`o2-mgr` is a simple program to bootstrap and manage a development environment for the [ALICE-O2](https://github.com/AliceO2Group/AliceO2) project.

## Example

```sh
$> cd my-dev
$> o2-mgr 
o2-mgr: bootstrapping O2 into [/home/binet/dev/alice/o2/o2-mgr/my-dev]...
Cloning into '/home/binet/dev/alice/o2/o2-mgr/my-dev/src/fair-soft'...
Cloning into '/home/binet/dev/alice/o2/o2-mgr/my-dev/src/fair-root'...
Cloning into '/home/binet/dev/alice/o2/o2-mgr/my-dev/src/alice-o2'...
o2-mgr: using config:
===
compiler=gcc
debug=no
optimize=yes
geant4_download_install_data_automatic=no
geant4_install_data_from_dir=no
build_root6=yes
build_python=no
install_sim=yes
SIMPATH_INSTALL=/home/binet/dev/alice/o2/o2-mgr/my-dev/rootfs/opt/alice/sw/externals
platform=linux
===

o2-mgr: building fair-soft...
The build process for the external packages for the FairRoot Project was started at 090316_100750
*** Compiling the external packages with the GCC compiler
[... building FairSoft ...]

o2-mgr: building fair-root...
[... building FairRoot ...]

o2-mgr: building alice-o2...
[... building AliceO2  ...]

Scanning dependencies of target testits
[ 98%] Building CXX object test/its/CMakeFiles/testits.dir/HitAnalysis.cxx.o
[ 99%] Building CXX object test/its/CMakeFiles/testits.dir/G__testitsDict.cxx.o
[100%] Linking CXX shared library ../../lib/libtestits.so
[100%] Built target testits
o2-mgr: 

o2-mgr: ::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
o2-mgr: o2 build complete
o2-mgr: run:
$> source /opt/alice/src/alice-o2/build/config.sh

o2-mgr: to get a runtime environment.
o2-mgr: build complete.
o2-mgr: bye.
```

