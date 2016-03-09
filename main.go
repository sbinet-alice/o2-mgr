package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Build struct {
	Dir          string
	RootFS       string
	SimPath      string
	FairRootPath string
	Deps         []string
	Docker       bool
}

var cfg Build

func main() {
	flag.BoolVar(&cfg.Docker, "use-container", false, "switch to build inside container")

	flag.Parse()

	log.SetFlags(0)
	log.SetPrefix("o2-mgr ")

	dir := "."
	if flag.NArg() == 1 {
		dir = flag.Arg(0)
	}

	dir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("bootstrapping O2 into [%s]...\n", dir)
	err = os.MkdirAll(filepath.Join(dir, "rootfs", "opt", "alice"), 0755)
	if err != nil {
		log.Fatalf("error creating rootfs: %v\n", err)
	}

	cfg.Dir = dir
	cfg.RootFS = filepath.Join(dir, "rootfs")
	cfg.SimPath = "/opt/alice/sw/externals"
	cfg.FairRootPath = "/opt/alice/sw"
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

	if !cfg.Docker {
		cfg.SimPath = filepath.Join(cfg.RootFS, cfg.SimPath)
		cfg.FairRootPath = filepath.Join(cfg.RootFS, cfg.FairRootPath)
	}

	for _, v := range []struct {
		Repo string
		Dir  string
	}{
		{
			Repo: "git://github.com/FairRootGroup/FairSoft",
			Dir:  filepath.Join(dir, "src/fair-soft"),
		},
		{
			Repo: "git://github.com/FairRootGroup/FairRoot",
			Dir:  filepath.Join(dir, "src/fair-root"),
		},
		{
			Repo: "git://github.com/AliceO2Group/AliceO2",
			Dir:  filepath.Join(dir, "src/alice-o2"),
		},
	} {
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
build_python=no
install_sim=yes
SIMPATH_INSTALL=%s
platform=linux
`,
		cfg.SimPath,
	)

	log.Printf("using config:\n===%v\n===\n", config)
	err = ioutil.WriteFile(
		filepath.Join(dir, "src/fair-config.cache"),
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

	return err
}

func (b *Build) buildFairSoft() error {
	var err error
	cmd := command("/bin/sh", "-c", "./configure.sh ../fair-config.cache")
	cmd.Dir = filepath.Join(b.Dir, "src/fair-soft")
	err = cmd.Run()
	if err != nil {
		log.Printf("error running command %s %s: %v\n", cmd.Path, strings.Join(cmd.Args, " "), err)
		return err
	}
	return err
}

func (b *Build) buildFairRoot() error {
	var err error
	bdir := filepath.Join(b.Dir, "src", "fair-root", "build")
	err = os.MkdirAll(bdir, 0644)
	if err != nil {
		return err
	}

	for _, args := range [][]string{
		{"cmake", "-DCMAKE_INSTALL_PREFIX=" + b.FairRootPath, filepath.Join(b.Dir, "src/fair-root")},
		{"make"},
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
	bdir := filepath.Join(b.Dir, "src", "alice-o2", "build")
	err = os.MkdirAll(bdir, 0644)
	if err != nil {
		return err
	}

	for _, args := range [][]string{
		{"cmake", "../"},
		{"make"},
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
	log.Printf("to get a runtime environment\n")

	return err
}

func command(args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = cfg.RootFS
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
