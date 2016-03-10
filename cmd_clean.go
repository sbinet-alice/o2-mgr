package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func oxyMakeClean() *commander.Command {
	cmd := &commander.Command{
		Run:       oxyRunClean,
		UsageLine: "clean [options] [target-name [target-name [...]]]",
		Short:     "clean target(s)",
		Long: `
clean cleans out (remove) targets.

ex:
 $ oxy clean
 $ oxy clean root
 $ oxy clean root fair-soft fair-root
`,
		Flag: *flag.NewFlagSet("oxy-clean", flag.ExitOnError),
	}
	cmd.Flag.Bool("v", false, "enable verbose output")
	return cmd
}

func oxyRunClean(cmd *commander.Command, args []string) error {
	var err error

	if mgr.Container == "" {
		log.Printf("error: no container registered.\n")
		log.Printf("please run:\n\n\t$> %s init-container\n\n", os.Args[0])

		return fmt.Errorf("no container registered")
	}

	extpkgs, err := mgr.getExtPkgs()
	if err != nil {
		return err
	}

	tgts := make([]string, len(args))
	copy(tgts, args)

	if len(tgts) == 0 {
		tgts = append(tgts, "alice-o2", "fair-root", "fair-soft")
	}

	for _, arg := range args {
		switch arg {
		case "fair-soft":
			err = mgr.cleanFairSoft()
			if err != nil {
				return err
			}

		case "fair-root":
			err = mgr.cleanFairRoot()
			if err != nil {
				return err
			}

		case "alice-o2":
			err = mgr.cleanO2()
			if err != nil {
				return err
			}

		default:
			_, ok := extpkgs[arg]
			if !ok {
				log.Printf("unknown target [%s]\n", arg)
				return fmt.Errorf("unknown target [%s]", arg)
			}
			err = mgr.cleanExtPkg(arg)
			if err != nil {
				return err
			}
		}
	}

	return err
}
