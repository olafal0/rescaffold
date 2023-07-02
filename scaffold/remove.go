package scaffold

import (
	"fmt"
	"os"
	"path"

	"github.com/olafal0/rescaffold/config"
)

func Remove(lockfile *config.Lockfile, scaffoldSource, outdir string) error {
	// TODO: load scaffold from git or whatever if name is a URL
	// For now, assume source == name == local directory
	scaffoldDir := scaffoldSource
	scaf, err := LoadFromDir(scaffoldDir)
	if err != nil {
		return err
	}

	lockedScaffold := lockfile.GetScaffold(scaffoldDir, scaffoldSource, scaf.Manifest)

	// Find all vars in the manifest
	// If any do not have values in the lockfile, prompt the user for them
	varValues, err := LoadVarsInteractive(scaf.Manifest.Vars, lockedScaffold.Vars)
	if err != nil {
		return err
	}
	lockedScaffold.Vars = varValues

	replacer, err := RegexpLoopReplacer(scaf.Manifest, varValues)
	if err != nil {
		return err
	}

	for _, scaffoldFile := range scaf.Files {
		outFilename := replacer(scaffoldFile.RelativePath)
		outpath := path.Join(outdir, outFilename)

		// Get file info from lockfile (may be nil)
		lockedFile := lockedScaffold.GetFile(outpath)

		// Check if file exists at destination
		existingOutfile, err := os.Open(outpath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error checking if file exists: %w", err)
		}

		if existingOutfile == nil {
			// File does not exist. If it's not in the lockfile, nothing to do
			if lockedFile == nil {
				continue
			}

			// Otherwise, remove the file from the lockfile
			lockedScaffold.RemoveFile(outpath)
			if err := lockfile.WriteUpdated(); err != nil {
				return fmt.Errorf("error writing updated lockfile: %w", err)
			}
			continue
		}

		// If file exists, check that its contents are what we expect (matching checksum)
		if lockedFile == nil {
			return fmt.Errorf("file exists but is not in lockfile: %s", outpath)
		}

		checksum, err := hashFile(existingOutfile)
		if err != nil {
			return err
		}

		if lockedFile.Checksum != checksum {
			fmt.Printf("file has been modified, skipping: %s\n", outpath)
			continue
		}

		// Output file exists and matches expected value. We're removing, so this
		// file should be deleted

		if err := os.Remove(outpath); err != nil {
			return fmt.Errorf("error removing file: %w", err)
		}

		// Update lockfile with new file information for each new file
		lockedScaffold.RemoveFile(outpath)
		if err := lockfile.WriteUpdated(); err != nil {
			return fmt.Errorf("error writing updated lockfile: %w", err)
		}
	}

	lockfile.RemoveScaffold(scaffoldDir)
	if len(lockfile.Scaffolds) > 0 {
		if err := lockfile.WriteUpdated(); err != nil {
			return fmt.Errorf("error writing updated lockfile: %w", err)
		}
	} else {
		if err := lockfile.Remove(); err != nil {
			return fmt.Errorf("error removing lockfile: %w", err)
		}
	}

	// Walk outdir and remove any empty directories
	_, err = removeEmptyDirectories("", outdir)
	if err != nil {
		return err
	}
	return nil
}

func removeEmptyDirectories(basePath, dir string) (removed bool, err error) {
	files, err := os.ReadDir(path.Join(basePath, dir))
	if err != nil {
		return false, err
	}

	encounteredNonDir := false
	for _, file := range files {
		if file.IsDir() {
			childRemoved, err := removeEmptyDirectories(path.Join(basePath, dir), file.Name())
			if err != nil {
				return false, err
			}
			if !childRemoved {
				encounteredNonDir = true
			}
		} else {
			encounteredNonDir = true
		}
	}

	if basePath == "" {
		return false, nil
	}
	if !encounteredNonDir {
		return true, os.Remove(path.Join(basePath, dir))
	}
	return false, nil
}
