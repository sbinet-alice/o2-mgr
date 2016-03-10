package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func oxyMakeInit() *commander.Command {
	cmd := &commander.Command{
		Run:       oxyRunInit,
		UsageLine: "init [options] <workarea>",
		Short:     "initialize a new workarea",
		Long: `
init initializes a new workarea.

ex:
 $ oxy init
 $ oxy init .
 $ oxy init my-work-area
`,
		Flag: *flag.NewFlagSet("oxy-init", flag.ExitOnError),
	}
	cmd.Flag.Bool("v", false, "enable verbose output")
	return cmd
}

func oxyRunInit(cmd *commander.Command, args []string) error {
	var err error
	dir := ""

	switch len(args) {
	case 0:
		dir = "."
	case 1:
		dir = args[0]
	default:
		return fmt.Errorf("you need to give a directory name")
	}

	dir = os.ExpandEnv(dir)
	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}
	dir = filepath.Clean(dir)

	log.Printf("init [%s]...\n", dir)
	mgr.Dir = dir
	mgr.SrcDir = filepath.Join(dir, "src")
	mgr.SimPath = filepath.Join(dir, "externals")
	mgr.FairRootPath = filepath.Join(dir, "install")

	for _, dir := range []string{
		mgr.SrcDir,
		mgr.SimPath,
		mgr.FairRootPath,
	} {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	for _, v := range []struct {
		Repo string
		Dir  string
	}{
		{
			Repo: "git://github.com/FairRootGroup/FairSoft",
			Dir:  filepath.Join(mgr.SrcDir, "fair-soft"),
		},
		{
			Repo: "git://github.com/FairRootGroup/FairRoot",
			Dir:  filepath.Join(mgr.SrcDir, "fair-root"),
		},
		{
			Repo: "git://github.com/AliceO2Group/AliceO2",
			Dir:  filepath.Join(mgr.SrcDir, "alice-o2"),
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

	var config = `## config for FairRoot externals
compiler=gcc
debug=no
optimize=yes
geant4_download_install_data_automatic=no
geant4_install_data_from_dir=no
build_root6=yes
build_python=yes
install_sim=yes
SIMPATH_INSTALL=/opt/alice/sw/externals
platform=linux
`

	cfgname := filepath.Join(mgr.SrcDir, "config.cache")
	log.Printf("config [%s]:\n%v", cfgname, config)
	err = ioutil.WriteFile(cfgname, []byte(config), 0644)
	if err != nil {
		return err
	}

	cfg, err := os.Create(filepath.Join(mgr.Dir, oxyCache))
	if err != nil {
		return err
	}
	defer cfg.Close()
	cfgdata, err := json.MarshalIndent(mgr, "", "  ")
	if err != nil {
		return err
	}
	_, err = cfg.Write(cfgdata)
	if err != nil {
		return err
	}
	err = cfg.Close()
	if err != nil {
		return err
	}

	c := mgr.command("oxy", "init-container")
	c.Dir = mgr.Dir
	err = c.Run()
	if err != nil {
		return err
	}

	return err
}
