package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
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
	Env          struct {
		CXX string
		CC  string
	}

	mux   sync.RWMutex
	docks map[string]struct{} // containers spawned on our behalf
}

func newMgr() *Mgr {
	mgr := &Mgr{
		docks: make(map[string]struct{}),
	}
	mgr.Env.CXX = "/usr/bin/g++"
	mgr.Env.CC = "/usr/bin/gcc"

	go func() {
		sigch := make(chan os.Signal)
		signal.Notify(sigch, os.Interrupt, os.Kill)
		for {
			select {
			case <-sigch:
				os.Stderr.Sync()
				os.Stdout.Sync()
				mgr.Release()
				os.Stderr.Sync()
				os.Stdout.Sync()
			}
		}
	}()

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("error inferring current working directory: %v", err)
	}

	f, err := os.Open(filepath.Join(pwd, oxyCache))
	if err != nil {
		return mgr
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(mgr)
	if err != nil {
		log.Printf("error loading oxy config from cache [%s]: %v\n", f.Name(), err)
		return mgr
	}
	return mgr
}

func (mgr *Mgr) Release() {
	mgr.mux.RLock()
	defer mgr.mux.RUnlock()

	for id := range mgr.docks {
		mgr.rmContainer(id)
	}
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

	err = mgr.buildO2()
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
	err = os.MkdirAll(bdir, 0755)
	if err != nil {
		return err
	}

	for _, args := range [][]string{
		{
			"cmake", "-DCMAKE_INSTALL_PREFIX=" + container.FairRootPath,
			"-DUSE_NANOMSG=1",
			"-DCMAKE_CXX_COMPILER=" + mgr.Env.CXX,
			"-DCMAKE_CC_COMPILER=" + mgr.Env.CC,
			"../",
		},
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

func (mgr *Mgr) buildO2() error {
	var err error
	log.Printf("building alice-o2...\n")
	bdir := filepath.Join(mgr.SrcDir, "alice-o2", "build")
	err = os.MkdirAll(bdir, 0755)
	if err != nil {
		return err
	}

	for _, args := range [][]string{
		{"cmake", "../"},
		{"make", fmt.Sprintf("-j%d", runtime.NumCPU())},
	} {
		cmd := dockerCmd{Cmd: args[0], Args: args[1:], Dir: container.SrcDir + "/alice-o2/build"}
		err = mgr.drun(cmd)
		if err != nil {
			log.Printf("error running command %s: %v\n", cmd, err)
			return err
		}
	}

	log.Printf("\n\n")
	log.Printf(strings.Repeat(":", 80))
	log.Printf("oxy build complete\n")
	log.Printf("run:\n$> source %s/config.sh\n\n", container.SrcDir+"/alice-o2/build")
	log.Printf("to get a runtime environment.\n")
	return err
}

func (mgr *Mgr) getExtPkgs() (map[string]int, error) {
	pkgs := make(map[string]int)
	f, err := os.Open(filepath.Join(mgr.SrcDir, "fair-soft", "make_clean.sh"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	re := regexp.MustCompile(`^clean_(?P<PackageName>.*?)\(\).*?\{.*?`)
	for scan.Scan() {
		err = scan.Err()
		if err != nil {
			break
		}
		txt := scan.Text()
		if !re.MatchString(txt) {
			continue
		}
		sub := re.FindStringSubmatch(txt)
		pkg := sub[1]
		if pkg == "all" {
			continue
		}
		pkgs[pkg] = 1

	}
	err = scan.Err()
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	return pkgs, err
}

func (mgr *Mgr) cleanExtPkg(pkg string) error {
	cmd := dockerCmd{Cmd: "./make_clean.sh", Args: []string{pkg}, Dir: container.SrcDir + "/fair-soft/build"}
	return mgr.drun(cmd)
}

func (mgr *Mgr) cleanFairSoft() error {
	return mgr.cleanExtPkg("all")
}

func (mgr *Mgr) cleanFairRoot() error {
	cmd := dockerCmd{Cmd: "make", Args: []string{"clean"}, Dir: container.SrcDir + "/fair-root/build"}
	return mgr.drun(cmd)
}

func (mgr *Mgr) cleanO2() error {
	cmd := dockerCmd{Cmd: "make", Args: []string{"clean"}, Dir: container.SrcDir + "/alice-o2/build"}
	return mgr.drun(cmd)
}

func (mgr *Mgr) command(cmd string, args ...string) *exec.Cmd {
	c := exec.Command(cmd, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c
}

func (mgr *Mgr) drun(cmd dockerCmd) error {
	// id := fmt.Sprintf("oxy-dev-%d", time.Now().Unix())
	id := "oxy-box"
	mgr.addContainer(id)
	defer mgr.rmContainer(id)

	dargs := []string{
		"run",
		"-it", "--rm",
		"-v", fmt.Sprintf("%s:%s", mgr.SrcDir, container.SrcDir),
		"-v", fmt.Sprintf("%s:%s", mgr.SimPath, container.SimPath),
		"-v", fmt.Sprintf("%s:%s", mgr.FairRootPath, container.FairRootPath),
		"-u", fmt.Sprintf("%s:%s", container.Uid, container.Uid),
		"-w", cmd.Dir,
		"-e", "SIMPATH=" + container.SimPath,
		"-e", "FAIRROOTPATH=" + container.FairRootPath,
		"-e", "CXX=" + mgr.Env.CXX,
		"-e", "CC=" + mgr.Env.CC,
		"--name", id,
		"-h", id,
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

func (mgr *Mgr) addContainer(id string) {
	mgr.mux.Lock()
	mgr.docks[id] = struct{}{}
	mgr.mux.Unlock()
}

func (mgr *Mgr) rmContainer(id string) {
	mgr.mux.Lock()
	exec.Command("docker", "kill", id).Run()
	exec.Command("docker", "rm", id).Run()
	delete(mgr.docks, id)
	mgr.mux.Unlock()
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
	Uid          string
	SrcDir       string
	SimPath      string
	FairRootPath string
}{
	Uid:          "oxy",
	SrcDir:       "/opt/alice/src",
	SimPath:      "/opt/alice/sw/externals",
	FairRootPath: "/opt/alice/sw/install",
}
