package scaffold

import (
	"errors"
	"os"
	"path"
	"strings"

	"github.com/olafal0/rescaffold/config"
)

type ScaffoldFile struct {
	// RelativePath is the file path relative to the root of the scaffold directory
	RelativePath string
	// FullPath is the absolute path to the file
	FullPath string
}

type Scaffold struct {
	Files    []ScaffoldFile
	Manifest *config.Manifest
	Vars     map[string]string
}

func LoadFromDir(dirName string) (*Scaffold, error) {
	filenames, err := walkDir("", dirName)
	if err != nil {
		return nil, err
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	scaffold := &Scaffold{}
	scaffold.Files = make([]ScaffoldFile, 0, len(filenames))
	for _, filename := range filenames {
		if path.Base(filename) == config.ManifestFilename {
			// Read and load manifest file
			f, err := os.Open(filename)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			manifest, err := config.ParseManifest(f)
			if err != nil {
				return nil, err
			}
			scaffold.Manifest = manifest
			// Don't add manifest to scaffold files
			continue
		}

		scafFile := ScaffoldFile{
			RelativePath: path.Clean(strings.TrimPrefix(filename, dirName)),
		}
		if path.IsAbs(filename) {
			scafFile.FullPath = filename
		} else {
			scafFile.FullPath = path.Join(wd, filename)
		}
		scaffold.Files = append(scaffold.Files, scafFile)
	}
	if scaffold.Manifest == nil {
		return nil, errors.New("scaffold directory does not contain a manifest file")
	}
	return scaffold, nil
}

func walkDir(basePath, dir string) ([]string, error) {
	files, err := os.ReadDir(path.Join(basePath, dir))
	if err != nil {
		return nil, err
	}

	filePaths := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			subFiles, err := walkDir(path.Join(basePath, dir), file.Name())
			if err != nil {
				return nil, err
			}
			filePaths = append(filePaths, subFiles...)
		} else {
			filePaths = append(filePaths, path.Join(basePath, dir, file.Name()))
		}
	}
	return filePaths, nil
}
