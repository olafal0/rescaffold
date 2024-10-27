package scaffold

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/olafal0/rescaffold/config"
)

func Generate(lockfile *config.Lockfile, scaffoldSource, outdir string) error {
	scaf, err := LoadScaffold(scaffoldSource)
	if err != nil {
		return err
	}
	defer scaf.Cleanup()

	lockedScaffold := lockfile.GetScaffold(scaffoldSource, scaf.Manifest)

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

		// If file exists, check that its contents are what we expect (matching checksum)
		if existingOutfile != nil {
			if lockedFile == nil {
				fmt.Printf("file already exists but is not in lockfile, skipping: %s\n", outpath)
				continue
			}

			checksum, err := hashFile(existingOutfile)
			if err != nil {
				return err
			}

			if lockedFile.Checksum != checksum {
				return fmt.Errorf("file has been modified: %s", outpath)
			}

			// Output file exists and matches expected value, continue
			continue
		}

		// File does not exist; create directories in path and create file for writing
		err = os.MkdirAll(path.Dir(outpath), 0755)
		if err != nil {
			return fmt.Errorf("error creating subdirectories: %w", err)
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

	if scaf.Manifest.Meta.PostInstall != "" {
		fmt.Println("Post-install instructions:")
		fmt.Println(scaf.Manifest.Meta.PostInstall)
	}
	return nil
}

func ApplyTemplate(src io.Reader, dst io.Writer, replacer func(string) string) (checksum string, err error) {
	hasher := sha256.New()
	srcScan := bufio.NewScanner(src)
	for srcScan.Scan() {
		line := srcScan.Text()
		replacedLine := append([]byte(replacer(line)), '\n')
		_, err := dst.Write(replacedLine)
		if err != nil {
			return "", err
		}
		_, err = hasher.Write(replacedLine)
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// hashFile returns the sha256 hash of the contents of the given file. hashFile
// seeks to the beginning of the file before returning.
func hashFile(f io.ReadSeeker) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", fmt.Errorf("error hashing file: %w", err)
	}

	if _, err := f.Seek(0, 0); err != nil {
		return "", fmt.Errorf("error hashing file (seek): %w", err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
