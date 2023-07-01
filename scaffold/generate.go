package scaffold

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/olafal0/rescaffold/config"
)

func Generate(ctx context.Context, scaffoldSource string) error {
	outdir, ok := config.CtxOutputDir(ctx)
	if !ok {
		outdir = "."
	}
	lockfile, ok := config.CtxLockfile(ctx)
	if !ok {
		return errors.New("no lockfile loaded")
	}

	// TODO: load scaffold from git or whatever if name is a URL
	// For now, assume source == name == local directory
	scaffoldDir := scaffoldSource
	scaf, err := LoadFromDir(scaffoldDir)
	if err != nil {
		return err
	}
	defer scaf.Close()

	lockedScaffold := lockfile.GetScaffold(scaffoldDir, scaffoldSource, scaf.Manifest)

	// Find all vars in the manifest
	// If any do not have values in the lockfile, prompt the user for them
	varValues := make(map[string]string, len(scaf.Manifest.Vars))
	for varName, varOptions := range scaf.Manifest.Vars {
		if _, ok := lockedScaffold.Vars[varName]; ok {
			varValues[varName] = lockedScaffold.Vars[varName]
			continue
		}
		fmt.Printf("%s: %s\n", varName, varOptions.Description)
		if varOptions.Default != "" {
			fmt.Printf("Enter value [%s]: ", varOptions.Default)
		} else {
			fmt.Println("Enter value: ")
		}
		varValue := ""

		stdinScanner := bufio.NewScanner(os.Stdin)
		stdinScanner.Scan()
		varValue = stdinScanner.Text()
		if varValue == "" && varOptions.Default != "" {
			varValue = varOptions.Default
		}
		if varValue == "" && varOptions.Default == "" {
			return fmt.Errorf("var %s is required", varName)
		}
		varValues[varName] = varValue
	}

	lockedScaffold.Vars = varValues

	replacer := Replacer(scaf.Manifest, varValues)

	for _, scafFile := range scaf.Files {
		outFilename := replacer(scafFile.Path)
		outpath := path.Join(outdir, outFilename)

		// Get file info from lockfile (may be nil)
		lockedFile := lockedScaffold.GetFile(outpath)

		// Check if file exists at destination
		outf, err := os.Open(outpath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error checking if file exists: %w", err)
		}
		outfileExists := err == nil

		// If file does not exist, create directories in path and create file for writing
		if !outfileExists {
			err = os.MkdirAll(path.Dir(outpath), 0755)
			if err != nil {
				return fmt.Errorf("error creating subdirectories: %w", err)
			}
			outf, err = os.Create(outpath)
			if err != nil {
				return fmt.Errorf("error creating file: %w", err)
			}
		}

		// If file exists, check that its contents are what we expect (matching checksum)
		if outfileExists {
			if lockedFile == nil {
				return fmt.Errorf("file already exists but is not in lockfile: %s", outpath)
			}

			checksum, err := hashFile(outf)
			if err != nil {
				return err
			}

			if lockedFile.Checksum != checksum {
				return fmt.Errorf("file has been modified: %s", outpath)
			}

			// Output file exists and matches expected value, continue
			continue
		}

		// Stream file from input -> template replacement -> output
		err = ApplyTemplate(scafFile.Contents, outf, replacer)
		if err != nil {
			return fmt.Errorf("error applying template: %w", err)
		}

		if _, err := outf.Seek(0, 0); err != nil {
			return fmt.Errorf("error seeking in output file: %w", err)
		}

		checksum, err := hashFile(outf)
		if err != nil {
			return err
		}

		// Update lockfile with new file information for each new file
		lockedScaffold.SetFile(outpath, checksum)
		if err := lockfile.WriteUpdated(); err != nil {
			return fmt.Errorf("error writing updated lockfile: %w", err)
		}
	}
	return nil
}

func ApplyTemplate(in io.Reader, out io.Writer, replacer func(string) string) error {
	inScan := bufio.NewScanner(in)
	for inScan.Scan() {
		line := inScan.Text()
		_, err := out.Write(append([]byte(replacer(line)), '\n'))
		if err != nil {
			return err
		}
	}

	return nil
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
