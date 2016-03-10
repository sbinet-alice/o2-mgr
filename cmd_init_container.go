package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func oxyMakeInitContainer() *commander.Command {
	cmd := &commander.Command{
		Run:       oxyRunInitContainer,
		UsageLine: "init-container [options] <container-name>",
		Short:     "init a container",
		Long: `
init-container creates a new container to be used for compilation.

ex:
 $ oxy init-container
 $ oxy init-container my-dev
`,
		Flag: *flag.NewFlagSet("oxy-init-container", flag.ExitOnError),
	}
	cmd.Flag.Bool("v", false, "enable verbose output")
	return cmd
}

func oxyRunInitContainer(cmd *commander.Command, args []string) error {
	var err error
	cname := ""

	switch len(args) {
	case 0:
		cname = "oxy-dev"
	case 1:
		cname = args[0]
	default:
		return fmt.Errorf("you need to give a container name")
	}

	log.Printf("init-container [%s]...\n", cname)
	mgr.Container = cname

	err = mgr.saveJSON(filepath.Join(mgr.Dir, oxyCache))
	if err != nil {
		return err
	}

	log.Printf("retrieving latest centos:7 image...\n")
	err = mgr.command("docker", "pull", "centos:7").Run()
	if err != nil {
		return err
	}

	log.Printf("building container [%s]...\n", mgr.Container)
	tmp, err := ioutil.TempDir("", "oxy-docker-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	deps := []string{
		"cmake", "gcc", "gcc-c++", "gcc-gfortran", "make", "patch", "sed",
		"libX11-devel", "libXft-devel", "libXpm-devel", "libXext-devel",
		"libXmu-devel", "mesa-libGLU-devel", "mesa-libGL-devel", "ncurses-devel",
		"curl", "bzip2", "libbz2-dev", "gzip", "unzip", "tar", "expat-devel",
		"subversion", "git", "flex", "bison", "imake", "redhat-lsb-core",
		"python-devel", "libxml2-devel", "wget", "openssl-devel", "curl-devel",
		"automake", "autoconf", "libtool", "which",
	}

	usr, err := user.Current()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(
		filepath.Join(tmp, "Dockerfile"),
		[]byte(fmt.Sprintf(`## Dockerfile for alice-oxy base dev env.
from centos:7

run yum update -y && yum install -y %[1]s

## create a user for the container
run useradd -m -u %[2]s oxy-dev
user oxy-dev
	`, strings.Join(deps, " "), usr.Uid)),
		0644,
	)
	if err != nil {
		return err
	}

	c := mgr.command("docker", "build", "--rm", "-t", mgr.Container, ".")
	c.Dir = tmp
	err = c.Run()
	if err != nil {
		return err
	}

	return err
}
