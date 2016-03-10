package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	oxyCache = ".oxy.json"
)

type Mgr struct {
	Dir          string
	SrcDir       string
	SimPath      string
	FairRootPath string
	Deps         []string
	Container    string
}

func newMgr() *Mgr {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("error inferring current working directory: %v", err)
	}

	f, err := os.Open(filepath.Join(pwd, oxyCache))
	if err != nil {
		return &Mgr{}
	}
	defer f.Close()

	var mgr Mgr
	err = json.NewDecoder(f).Decode(&mgr)
	if err != nil {
		log.Printf("error loading oxy config from cache [%s]: %v\n", f.Name(), err)
		return &Mgr{}
	}
	return &mgr
}

func (mgr *Mgr) saveJSON(fname string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.MarshalIndent(mgr, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func (mgr *Mgr) Build() error {
	var err error

	for _, v := range []struct {
		key string
		val string
	}{
		{
			key: "SIMPATH",
			val: mgr.SimPath,
		},
		{
			key: "FAIRROOTPATH",
			val: mgr.FairRootPath,
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

	err = mgr.buildFairSoft()
	if err != nil {
		return err
	}

	err = mgr.buildFairRoot()
	if err != nil {
		return err
	}

	err = mgr.buildoxy()
	if err != nil {
		return err
	}

	log.Printf("build complete.\n")
	log.Printf("bye.\n")
	return err
}

func (mgr *Mgr) buildFairSoft() error {
	log.Printf("building fair-soft...\n")
	cmd := dockerCmd{Cmd: "/bin/sh", Args: []string{"-c", "./configure.sh ../config.cache"}, Dir: container.SrcDir + "/fair-soft"}
	err := mgr.drun(cmd)
	if err != nil {
		log.Printf("error running command %s: %v\n", cmd, err)
		return err
	}
	return err
}

func (mgr *Mgr) buildFairRoot() error {
	var err error
	log.Printf("building fair-root...\n")
	bdir := filepath.Join(mgr.SrcDir, "fair-root", "build")
	err = os.MkdirAll(bdir, 0644)
	if err != nil {
		return err
	}

	for _, args := range [][]string{
		{"cmake", "-DCMAKE_INSTALL_PREFIX=" + container.FairRootPath, "-DUSE_NANOMSG=1", "../"},
		{"make", fmt.Sprintf("-j%d", runtime.NumCPU())},
		{"make", "install"},
	} {
		cmd := dockerCmd{Cmd: args[0], Args: args[1:], Dir: container.SrcDir + "/fair-root/build"}
		err = mgr.drun(cmd)
		if err != nil {
			log.Printf("error running command %s: %v\n", cmd, err)
			return err
		}
	}

	return err
}

func (mgr *Mgr) buildoxy() error {
	var err error
	log.Printf("building alice-oxy...\n")
	bdir := filepath.Join(mgr.SrcDir, "alice-oxy", "build")
	err = os.MkdirAll(bdir, 0644)
	if err != nil {
		return err
	}

	for _, args := range [][]string{
		{"cmake", "../"},
		{"make", fmt.Sprintf("-j%d", runtime.NumCPU())},
	} {
		cmd := dockerCmd{Cmd: args[0], Args: args[1:], Dir: container.SrcDir + "/alice-oxy/build"}
		err = mgr.drun(cmd)
		if err != nil {
			log.Printf("error running command %s: %v\n", cmd, err)
			return err
		}
	}

	log.Printf("\n\n")
	log.Printf(strings.Repeat(":", 80))
	log.Printf("oxy build complete\n")
	log.Printf("run:\n$> source %s/config.sh\n\n", bdir)
	log.Printf("to get a runtime environment.\n")
	return err
}

func (mgr *Mgr) command(cmd string, args ...string) *exec.Cmd {
	c := exec.Command(cmd, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c
}

func (mgr *Mgr) drun(cmd dockerCmd) error {
	dargs := []string{
		"run",
		"-v", fmt.Sprintf("%s:%s", mgr.SrcDir, container.SrcDir),
		"-v", fmt.Sprintf("%s:%s", mgr.SimPath, container.SimPath),
		"-v", fmt.Sprintf("%s:%s", mgr.FairRootPath, container.FairRootPath),
		"-u", "oxy-dev",
		"-w", cmd.Dir,
		"-e", "SIMPATH=" + container.SimPath,
		"-e", "FAIRROOTPATH=" + container.FairRootPath,
		"-e", "CXX=c++",
		"-e", "CC=cc",
		mgr.Container,
		cmd.Cmd,
	}

	dargs = append(dargs, cmd.Args...)

	c := exec.Command("docker", dargs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

type dockerCmd struct {
	Cmd  string
	Args []string
	Dir  string
}

func (cmd dockerCmd) String() string {
	str := cmd.Cmd
	if len(cmd.Args) > 0 {
		str += strings.Join(cmd.Args, " ")
	}
	return str
}

var container = struct {
	SrcDir       string
	SimPath      string
	FairRootPath string
}{
	SrcDir:       "/opt/alice/src",
	SimPath:      "/opt/alice/sw/externals",
	FairRootPath: "/opt/alice/sw/install",
}
