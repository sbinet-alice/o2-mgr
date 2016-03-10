package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

var (
	oxy *commander.Command
	mgr = newMgr()
)

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
	err := oxy.Flag.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	defer mgr.Release()

	args := oxy.Flag.Args()
	err = oxy.Dispatch(args)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	return
}

func command(args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = mgr.Dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func init() {
	log.SetPrefix("oxy: ")
	log.SetFlags(0)

	oxy = &commander.Command{
		UsageLine: "oxy manages an AliceO2 build and development environment",
		Short:     "manages a development environment",
		Subcommands: []*commander.Command{
			oxyMakeInit(),
			oxyMakeInitContainer(),
			oxyMakeBuild(),
			oxyMakeShell(),
		},
		Flag: *flag.NewFlagSet("oxy", flag.ExitOnError),
	}
}
