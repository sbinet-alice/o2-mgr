package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Build struct {
	Dir          string
	SrcDir       string
	SimPath      string
	FairRootPath string
	Deps         []string
}

var cfg Build

const deps = `
yum update -y
yum install -y cmake gcc gcc-c++ gcc-gfortran make patch sed \
  libX11-devel libXft-devel libXpm-devel libXext-devel \
  libXmu-devel mesa-libGLU-devel mesa-libGL-devel ncurses-devel \
  curl bzip2 libbz2-dev gzip unzip tar expat-devel \
  subversion git flex bison imake redhat-lsb-core \
  python-devel libxml2-devel wget openssl-devel curl-devel \
  automake autoconf libtool which
  `

func main() {
	flag.Parse()

	log.SetFlags(0)
	log.SetPrefix("o2-mgr: ")

	dir := "."
	if flag.NArg() == 1 {
		dir = flag.Arg(0)
	}

	dir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("bootstrapping O2 into [%s]...\n", dir)
	cfg.Dir = dir
	cfg.SrcDir = filepath.Join(dir, "src")
	cfg.SimPath = filepath.Join(dir, "sw/externals")
	cfg.FairRootPath = filepath.Join(dir, "sw")

	cfg.Deps = []string{
		"automake", "cmake", "cmake-data", "curl",
		"g++", "gcc", "gfortran",
		"build-essential", "make", "patch", "sed",
		"libcurl4-openssl-dev", "libtool",
		"libx11-dev", "libxft-dev",
		"libxext-dev", "libxpm-dev", "libxmu-dev", "libglu1-mesa-dev",
		"libgl1-mesa-dev", "ncurses-dev", "curl", "bzip2", "gzip", "unzip", "tar",
		"subversion", "git", "xutils-dev", "flex", "bison", "lsb-release",
		"python-dev", "libxml2-dev", "wget", "libssl-dev",
	}

	err = os.MkdirAll(cfg.Dir, 0755)
	if err != nil {
		log.Fatalf("error creating rootfs: %v\n", err)
	}

	for _, v := range []struct {
		Repo string
		Dir  string
	}{
		{
			Repo: "git://github.com/FairRootGroup/FairSoft",
			Dir:  filepath.Join(cfg.SrcDir, "fair-soft"),
		},
		{
			Repo: "git://github.com/FairRootGroup/FairRoot",
			Dir:  filepath.Join(cfg.SrcDir, "fair-root"),
		},
		{
			Repo: "git://github.com/AliceO2Group/AliceO2",
			Dir:  filepath.Join(cfg.SrcDir, "alice-o2"),
		},
	} {
		log.Printf("fetching repo [%s]...\n", v.Repo)
		_, err = os.Stat(v.Dir)
		if os.IsNotExist(err) {
			cmd := command("git", "clone", v.Repo, v.Dir)
			err = cmd.Run()
			if err != nil {
				log.Fatalf("error running %v: %v\n", cmd, err)
			}
			continue
		}
		cmd := command("git", "fetch", "--all")
		cmd.Dir = v.Dir
		err = cmd.Run()
		if err != nil {
			log.Fatalf("error running %v: %v\n", cmd, err)
		}
	}

	var config = fmt.Sprintf(
		`compiler=gcc
debug=no
optimize=yes
geant4_download_install_data_automatic=no
geant4_install_data_from_dir=no
build_root6=yes
build_python=yes
install_sim=yes
SIMPATH_INSTALL=%s
platform=linux
`,
		cfg.SimPath,
	)

	log.Printf("using config:\n===\n%v===\n\n", config)
	err = ioutil.WriteFile(
		filepath.Join(cfg.SrcDir, "fair-config.cache"),
		[]byte(config),
		0644,
	)
	if err != nil {
		log.Fatalf("error creating config-file: %v\n", err)
	}

	err = cfg.Build()
	if err != nil {
		log.Fatalf("error build: %v\n", err)
	}
}

func (b *Build) Build() error {
	var err error

	for _, v := range []struct {
		key string
		val string
	}{
		{
			key: "SIMPATH",
			val: b.SimPath,
		},
		{
			key: "FAIRROOTPATH",
			val: b.FairRootPath,
		},
		{
			key: "CC",
			val: "gcc",
		},
		{
			key: "CXX",
			val: "g++",
		},
	} {
		err = os.Setenv(v.key, v.val)
		if err != nil {
			log.Printf("error setenv %q to %q: %v\n", v.key, v.val, err)
			return err
		}

		err = os.Setenv("PATH", v.val+"/bin:"+os.Getenv("PATH"))
		if err != nil {
			log.Printf("error prepending %q to %q: %v\n", v.val+"/bin", "PATH", err)
			return err
		}

		err = os.Setenv("LD_LIBRARY_PATH", v.val+"/lib:"+os.Getenv("LD_LIBRARY_PATH"))
		if err != nil {
			log.Printf("error prepending %q to %q: %v\n", v.val+"/lib", "LD_LIBRARY_PATH", err)
			return err
		}
	}

	err = b.buildFairSoft()
	if err != nil {
		return err
	}

	err = b.buildFairRoot()
	if err != nil {
		return err
	}

	err = b.buildO2()
	if err != nil {
		return err
	}

	log.Printf("build complete.\n")
	log.Printf("bye.\n")
	return err
}

func (b *Build) buildFairSoft() error {
	var err error
	log.Printf("building fair-soft...\n")
	cmd := command("/bin/sh", "-c", "./configure.sh ../fair-config.cache")
	cmd.Dir = filepath.Join(b.SrcDir, "fair-soft")
	err = cmd.Run()
	if err != nil {
		log.Printf("error running command %s %s: %v\n", cmd.Path, strings.Join(cmd.Args, " "), err)
		return err
	}
	return err
}

func (b *Build) buildFairRoot() error {
	var err error
	log.Printf("building fair-root...\n")
	bdir := filepath.Join(b.SrcDir, "fair-root", "build")
	err = os.MkdirAll(bdir, 0644)
	if err != nil {
		return err
	}

	for _, args := range [][]string{
		{"cmake", "-DCMAKE_INSTALL_PREFIX=" + b.FairRootPath, "-DUSE_NANOMSG=1", filepath.Join(b.Dir, "src/fair-root")},
		{"make", fmt.Sprintf("-j%d", runtime.NumCPU())},
		{"make", "install"},
	} {
		cmd := command(args...)
		cmd.Dir = bdir
		err = cmd.Run()
		if err != nil {
			log.Printf("error running command %s %s: %v\n", cmd.Path, strings.Join(cmd.Args, " "), err)
			return err
		}
	}

	return err
}

func (b *Build) buildO2() error {
	var err error
	log.Printf("building alice-o2...\n")
	bdir := filepath.Join(b.SrcDir, "alice-o2", "build")
	err = os.MkdirAll(bdir, 0644)
	if err != nil {
		return err
	}

	for _, args := range [][]string{
		{"cmake", "../"},
		{"make", fmt.Sprintf("-j%d", runtime.NumCPU())},
	} {
		cmd := command(args...)
		cmd.Dir = bdir
		err = cmd.Run()
		if err != nil {
			log.Printf("error running command %s %s: %v\n", cmd.Path, strings.Join(cmd.Args, " "), err)
			return err
		}
	}

	log.Printf("\n\n")
	log.Printf(strings.Repeat(":", 80))
	log.Printf("o2 build complete\n")
	log.Printf("run:\n$> source %s/config.sh\n\n", bdir)
	log.Printf("to get a runtime environment.\n")
	return err
}

func command(args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = cfg.Dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
