package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func oxyMakeBuild() *commander.Command {
	cmd := &commander.Command{
		Run:       oxyRunBuild,
		UsageLine: "build [options] [target-name [target-name [...]]]",
		Short:     "build target(s)",
		Long: `
build builds targets.

ex:
 $ oxy build
 $ oxy build root
 $ oxy build root fair-soft fair-root
`,
		Flag: *flag.NewFlagSet("oxy-build", flag.ExitOnError),
	}
	cmd.Flag.Bool("v", false, "enable verbose output")
	return cmd
}

func oxyRunBuild(cmd *commander.Command, args []string) error {
	var err error

	if mgr.Container == "" {
		log.Printf("error: no container registered for building.\n")
		log.Printf("please run:\n\n\t$> %s init-container\n\n", os.Args[0])

		return fmt.Errorf("no container registered")
	}

	err = mgr.Build()
	if err != nil {
		return err
	}

	return err
}
