package scaffold

import (
	"errors"
	"io"
	"os"
	"path"
	"strings"

	"github.com/olafal0/rescaffold/config"
)

type ScaffoldFile struct {
	Path     string
	Contents io.ReadCloser
}

type Scaffold struct {
	Files    []ScaffoldFile
	Manifest *config.Manifest
	Vars     map[string]string
}

func LoadFromDir(dirName string) (*Scaffold, error) {
	files, err := walkDir("", dirName)
	if err != nil {
		return nil, err
	}

	scaffold := &Scaffold{}
	scaffold.Files = make([]ScaffoldFile, 0, len(files))
	for _, file := range files {
		// Open file for reading
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}

		if path.Base(file) == config.ManifestFilename {
			manifest, err := config.ParseManifest(f)
			if err != nil {
				return nil, err
			}
			scaffold.Manifest = manifest
			// Don't add manifest to scaffold files
			continue
		}

		scaffold.Files = append(scaffold.Files, ScaffoldFile{
			Path:     path.Clean(strings.TrimPrefix(file, dirName)),
			Contents: f,
		})
	}

	if scaffold.Manifest == nil {
		return nil, errors.New("scaffold directory does not contain a manifest file")
	}
	return scaffold, nil
}

func (s *Scaffold) Close() error {
	if s == nil {
		return nil
	}
	for _, file := range s.Files {
		if err := file.Contents.Close(); err != nil {
			return err
		}
	}
	return nil
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
