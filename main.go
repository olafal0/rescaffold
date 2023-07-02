package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"path"

	"github.com/olafal0/rescaffold/config"
	"github.com/olafal0/rescaffold/scaffold"
	"github.com/olafal0/rescaffold/set"
)

type Action int

const (
	ActionUpgrade Action = iota
	ActionRemove
	ActionGenerate
)

func GetAction(shouldUpgrade, shouldRemove bool) Action {
	switch {
	case shouldUpgrade:
		return ActionUpgrade
	case shouldRemove:
		return ActionRemove
	default:
		return ActionGenerate
	}
}

func main() {
	var shouldUpgrade, shouldRemove bool
	var outputDir string
	flag.BoolVar(&shouldUpgrade, "upgrade", false, "upgrade specified scaffolds, or all scaffolds if none are specified")
	flag.BoolVar(&shouldRemove, "remove", false, "remove specified scaffolds from the project")
	flag.StringVar(&outputDir, "out", ".", "directory in which scaffold files are placed")
	needHelp := flag.Bool("help", false, "print usage information")
	flag.Parse()
	scaffolds := flag.Args()

	if shouldUpgrade && shouldRemove {
		fmt.Println("Cannot upgrade and remove at the same time")
		flag.Usage()
		return
	}

	if *needHelp || (!shouldUpgrade && !shouldRemove && len(scaffolds) == 0) {
		flag.Usage()
		return
	}

	lockfilePath := path.Join(outputDir, config.LockfileFilename)
	lockfile, err := config.LoadLockfile(lockfilePath)
	if err != nil {
		log.Fatal(fmt.Errorf("could not load lockfile: %w", err))
	}

	switch {
	case shouldUpgrade:
		err = UpgradeScaffolds(lockfile, scaffolds, outputDir)
	case shouldRemove:
		err = RemoveScaffolds(lockfile, scaffolds, outputDir)
	default:
		err = GenerateScaffolds(lockfile, scaffolds, outputDir)
	}
	if err != nil {
		if lockfile.IsNewlyCreated() {
			lockfile.Remove()
		}
		log.Fatal(err)
	}
}

func UpgradeScaffolds(lockfile *config.Lockfile, scaffolds []string, outdir string) error {
	if len(scaffolds) == 0 {
		scaffolds = set.Keys(lockfile.Scaffolds)
	}
	for _, s := range scaffolds {
		err := scaffold.Upgrade(lockfile, path.Clean(s), outdir)
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveScaffolds(_ *config.Lockfile, scaffolds []string, _ string) error {
	if len(scaffolds) == 0 {
		return fmt.Errorf("will not remove all scaffolds without specifying them explicitly")
	}
	return errors.New("remove not yet implemented")
}

func GenerateScaffolds(lockfile *config.Lockfile, scaffolds []string, outdir string) error {
	if len(scaffolds) == 0 {
		return fmt.Errorf("cannot generate scaffolds if none are specified. to upgrade, use the -upgrade flag")
	}
	for _, s := range scaffolds {
		err := scaffold.Generate(lockfile, path.Clean(s), outdir)
		if err != nil {
			return err
		}
	}
	return nil
}
