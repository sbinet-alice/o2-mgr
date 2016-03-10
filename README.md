# oxy

`oxy` is a simple program to bootstrap and manage a development environment for the [ALICE-O2](https://github.com/AliceO2Group/AliceO2) project.

## Example

```sh
$> cd my-dev
$> oxy 
oxy: bootstrapping O2 into [/home/binet/dev/alice/o2/oxy/my-dev]...
Cloning into '/home/binet/dev/alice/o2/oxy/my-dev/src/fair-soft'...
Cloning into '/home/binet/dev/alice/o2/oxy/my-dev/src/fair-root'...
Cloning into '/home/binet/dev/alice/o2/oxy/my-dev/src/alice-o2'...
oxy: using config:
===
compiler=gcc
debug=no
optimize=yes
geant4_download_install_data_automatic=no
geant4_install_data_from_dir=no
build_root6=yes
build_python=no
install_sim=yes
SIMPATH_INSTALL=/home/binet/dev/alice/o2/oxy/my-dev/rootfs/opt/alice/sw/externals
platform=linux
===

oxy: building fair-soft...
The build process for the external packages for the FairRoot Project was started at 090316_100750
*** Compiling the external packages with the GCC compiler
[... building FairSoft ...]

oxy: building fair-root...
[... building FairRoot ...]

oxy: building alice-o2...
[... building AliceO2  ...]

Scanning dependencies of target testits
[ 98%] Building CXX object test/its/CMakeFiles/testits.dir/HitAnalysis.cxx.o
[ 99%] Building CXX object test/its/CMakeFiles/testits.dir/G__testitsDict.cxx.o
[100%] Linking CXX shared library ../../lib/libtestits.so
[100%] Built target testits
oxy: 

oxy: ::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
oxy: o2 build complete
oxy: run:
$> source /opt/alice/src/alice-o2/build/config.sh

oxy: to get a runtime environment.
oxy: build complete.
oxy: bye.
```

