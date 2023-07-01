package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"path"

	"github.com/olafal0/rescaffold/config"
	"github.com/olafal0/rescaffold/scaffold"
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

	ctx := context.Background()
	ctx = config.WithOutputDir(ctx, outputDir)

	lockfilePath := path.Join(outputDir, config.LockfileFilename)
	lockfile, err := config.LoadLockfile(lockfilePath)
	if err != nil {
		log.Fatal(fmt.Errorf("could not load lockfile: %w", err))
	}
	ctx = config.WithLockfile(ctx, lockfile)

	switch {
	case shouldUpgrade:
		err = UpgradeScaffolds(ctx, scaffolds)
	case shouldRemove:
		err = RemoveScaffolds(ctx, scaffolds)
	default:
		err = GenerateScaffolds(ctx, scaffolds)
	}
	if err != nil {
		if lockfile.IsNewlyCreated() {
			lockfile.Remove()
		}
		log.Fatal(err)
	}
}

func UpgradeScaffolds(_ context.Context, scaffolds []string) error {
	return errors.New("upgrade not yet implemented")
}

func RemoveScaffolds(_ context.Context, scaffolds []string) error {
	if len(scaffolds) == 0 {
		return fmt.Errorf("will not remove all scaffolds without specifying them explicitly")
	}
	return errors.New("remove not yet implemented")
}

func GenerateScaffolds(ctx context.Context, scaffolds []string) error {
	if len(scaffolds) == 0 {
		return fmt.Errorf("cannot generate scaffolds if none are specified. to upgrade, use the -upgrade flag")
	}
	for _, s := range scaffolds {
		err := scaffold.Generate(ctx, path.Clean(s))
		if err != nil {
			return err
		}
	}
	return nil
}
