package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func oxyMakeShell() *commander.Command {
	cmd := &commander.Command{
		Run:       oxyRunShell,
		UsageLine: "shell [options] [command [args...]]",
		Short:     "run a shell",
		Long: `
shell starts a command.

ex:
 $ oxy shell
 $ oxy shell ls /
`,
		Flag: *flag.NewFlagSet("oxy-shell", flag.ExitOnError),
	}
	cmd.Flag.Bool("v", false, "enable verbose output")
	return cmd
}

func oxyRunShell(cmd *commander.Command, args []string) error {
	var err error

	if mgr.Container == "" {
		log.Printf("error: no container registered\n")
		log.Printf("please run:\n\n\t$> %s init-container\n\n", os.Args[0])

		return fmt.Errorf("no container registered")
	}

	cargs := make([]string, len(args))
	copy(cargs, args)

	if len(cargs) == 0 {
		cargs = []string{"bash"}
	}

	err = mgr.drun(dockerCmd{Cmd: cargs[0], Args: cargs[1:], Dir: "/opt/alice"})
	if err != nil {
		return err
	}

	return err
}
