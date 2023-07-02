package scaffold

import (
	"fmt"
	"os"
	"path"

	"github.com/olafal0/rescaffold/config"
	"github.com/olafal0/rescaffold/set"
)

func Upgrade(lockfile *config.Lockfile, scaffoldSource, outdir string) error {
	// TODO: load scaffold from git or whatever if name is a URL
	// For now, assume source == name == local directory
	scaffoldDir := scaffoldSource
	scaf, err := LoadFromDir(scaffoldDir)
	if err != nil {
		return err
	}

	lockedScaffold := lockfile.GetScaffold(scaffoldDir, scaffoldSource, scaf.Manifest)

	// Create a set of output filenames that are present in the lockfile
	lockedFilePaths := set.NewWithCap[string](len(lockedScaffold.Files))
	for _, lockedFile := range lockedScaffold.Files {
		lockedFilePaths.Add(lockedFile.Path)
	}

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

		// Remove file from the set of locked filenames
		lockedFilePaths.Remove(outpath)

		// Check if file exists at destination
		existingOutfile, err := os.Open(outpath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error checking if file exists: %w", err)
		}

		// If file exists, check that its contents are what we expect (matching checksum)
		if existingOutfile != nil {
			if lockedFile == nil {
				return fmt.Errorf("file already exists but is not in lockfile: %s", outpath)
			}

			checksum, err := hashFile(existingOutfile)
			if err != nil {
				return err
			}

			if lockedFile.Checksum != checksum {
				fmt.Printf("file has been modified, skipping: %s\n", outpath)
				continue
			}

			// Output file exists and matches expected value. We're upgrading, so this
			// file may be rewritten
		} else {
			// File does not exist; create directories in path so we can open it for writing
			err = os.MkdirAll(path.Dir(outpath), 0755)
			if err != nil {
				return fmt.Errorf("error creating subdirectories: %w", err)
			}
		}

		outfile, err := os.Create(outpath)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}

		// Stream file from input -> template replacement -> output
		sourceFile, err := os.Open(scaffoldFile.FullPath)
		if err != nil {
			return fmt.Errorf("error opening source file: %w", err)
		}
		defer sourceFile.Close()

		newChecksum, err := ApplyTemplate(sourceFile, outfile, replacer)
		if err != nil {
			return fmt.Errorf("error applying template: %w", err)
		}

		// Update lockfile with new file information for each new file
		lockedScaffold.SetFile(outpath, newChecksum)
		if err := lockfile.WriteUpdated(); err != nil {
			return fmt.Errorf("error writing updated lockfile: %w", err)
		}
	}

	// Remove files that were present in the lockfile but not in the updated scaffold
	for lockedFilePath := range lockedFilePaths {
		lockedFile := lockedScaffold.GetFile(lockedFilePath)
		// Check that the file exists and checksum matches
		f, err := os.Open(lockedFilePath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("file present in lockfile, but could not open: %w", err)
		}
		if f == nil {
			// File does not exist, so it's already been deleted
			lockedScaffold.RemoveFile(lockedFilePath)
			if err := lockfile.WriteUpdated(); err != nil {
				return fmt.Errorf("error writing updated lockfile: %w", err)
			}
			continue
		}

		checksum, err := hashFile(f)
		if err != nil {
			return err
		}

		if lockedFile.Checksum != checksum {
			fmt.Printf("file should be deleted by upgrade, but has been modified - leaving in place: %s\n", lockedFilePath)
			// Leave the modified file in place but stop tracking it in the lockfile
			lockedScaffold.RemoveFile(lockedFilePath)
			if err := lockfile.WriteUpdated(); err != nil {
				return fmt.Errorf("error writing updated lockfile: %w", err)
			}

			continue
		}

		// Checksum matches, so the file exists and is unmodified, but the upgrade
		// removes it. Delete the file and remove it from the lockfile.
		if err := os.Remove(lockedFilePath); err != nil {
			return fmt.Errorf("error deleting file: %w", err)
		}
	}
	return nil
}
