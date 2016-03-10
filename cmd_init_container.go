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
		filepath.Join(tmp, "dot-bashrc"),
		[]byte(`
## .bashrc

if [ -f "/etc/bashrc" ] ; then
  source /etc/bashrc
fi

export PYTHONSTARTUP=$HOME/.pythonrc.py

## setup AliceO2 environment
if [ -e "/opt/alice/src/alice-o2/build/config.sh" ] ; then
   echo "::: sourcing alice-o2 environment..."
   . /opt/alice/src/alice-o2/build/config.sh
   export PATH=${FAIRROOTPATH}/bin:${PATH}
   export LD_LIBRARY_PATH=${FAIRROOTPATH}/lib:${SIMPATH}/lib:${LD_LIBRARY_PATH}
   echo "::: sourcing alice-o2 environment... [done]"
fi
`),
		0644,
	)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(
		filepath.Join(tmp, "dot-pythonrc.py"),
		[]byte(`
##
## for tab-completion
##
import rlcompleter, readline
readline.parse_and_bind('tab: complete')
readline.parse_and_bind( 'set show-all-if-ambiguous On' )

##
## for history
##
import os, atexit
histfile = os.path.join(os.environ["HOME"], ".python_history")
try:
    readline.read_history_file(histfile)
except IOError:
    pass
atexit.register(readline.write_history_file, histfile)
del os, atexit, histfile
del readline
`),
		0644,
	)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(
		filepath.Join(tmp, "Dockerfile"),
		[]byte(fmt.Sprintf(`## Dockerfile for alice-oxy base dev env.
from centos:7

run yum update -y && yum install -y %[1]s

## create a user for the container
run useradd -m -u %[2]s  -G wheel,root %[3]s
## add %[3]s to sudoers
run echo '%[3]s ALL=(ALL:ALL) ALL' >> /etc/sudoers

user %[3]s

env CC=%[4]s
env CXX=%[5]s

env USER=%[3]s
env HOME=/home/%[3]s

add dot-pythonrc.py $HOME/.pythonrc.py
add dot-bashrc      $HOME/.bashrc
`,
			strings.Join(deps, " "), usr.Uid, container.Uid,
			mgr.Env.CC,
			mgr.Env.CXX,
		)),
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
