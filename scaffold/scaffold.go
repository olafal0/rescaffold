package scaffold

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	giturls "github.com/chainguard-dev/git-urls"
	"github.com/google/uuid"
	"github.com/olafal0/rescaffold/config"
	"github.com/olafal0/rescaffold/set"
)

var (
	IgnoreFiles = set.New(".git")
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
	src      string
	cleanup  func() error
}

func (s *Scaffold) Cleanup() {
	if s.cleanup == nil {
		return
	}
	if err := s.cleanup(); err != nil {
		log.Printf("cleanup failed for %s: %v\n", s.src, err)
	}
}

func LoadScaffold(source string) (*Scaffold, error) {
	_, err := giturls.ParseScp(source)
	if err == nil {
		return LoadFromURL(source)
	}
	_, err = giturls.ParseTransport(source)
	if err == nil {
		return LoadFromURL(source)
	}
	return LoadFromDir(source)
}

func LoadFromDir(dirName string) (*Scaffold, error) {
	subScaffolds, err := findScaffolds(dirName, "", 0)
	if err != nil {
		return nil, err
	}
	switch len(subScaffolds) {
	case 0:
		return nil, fmt.Errorf("found no %s", config.ManifestFilename)
	case 1:
		if subScaffolds[0] != dirName {
			fmt.Printf("base directory is not a scaffold, using %s as the root\n", subScaffolds[0])
			dirName = subScaffolds[0]
		}
	default:
		return nil, fmt.Errorf("ambiguous scaffold reference, found sub-scaffolds: %v", subScaffolds)
	}

	filenames, err := walkDir(dirName, "")
	if err != nil {
		return nil, err
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	scaffold := &Scaffold{
		src: dirName,
	}
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

func LoadFromURL(source string) (*Scaffold, error) {
	destDir := "/tmp/rescaffold-" + uuid.NewString()
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("git clone %s %s", source, destDir),
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	cleanupFunc := func() error {
		if err := os.RemoveAll(destDir); err != nil {
			return err
		}
		return nil
	}

	scaf, err := LoadFromDir(destDir)
	if err != nil {
		log.Printf("failed to load cloned dir: %v\n", err)
		return nil, cleanupFunc()
	}
	scaf.src = source
	scaf.cleanup = cleanupFunc
	return scaf, nil
}

func isIgnored(f fs.DirEntry) bool {
	return IgnoreFiles.Contains(path.Base(f.Name()))
}

func walkDir(basePath, dir string) ([]string, error) {
	files, err := os.ReadDir(path.Join(basePath, dir))
	if err != nil {
		return nil, err
	}

	filePaths := make([]string, 0, len(files))
	for _, file := range files {
		if isIgnored(file) {
			continue
		}
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

const maxScaffoldSearchLevel = 1

func findScaffolds(basePath, dir string, level int) ([]string, error) {
	if level > maxScaffoldSearchLevel {
		return nil, nil
	}
	searchDir := path.Join(basePath, dir)
	files, err := os.ReadDir(searchDir)
	if err != nil {
		return nil, err
	}

	scaffoldDirs := []string{}
	for _, file := range files {
		if isIgnored(file) {
			continue
		}
		if file.Type().IsRegular() {
			if path.Base(file.Name()) == config.ManifestFilename {
				scaffoldDirs = append(scaffoldDirs, searchDir)
			}
		}
		if file.IsDir() {
			others, err := findScaffolds(searchDir, file.Name(), level+1)
			if err != nil {
				return nil, err
			}
			scaffoldDirs = append(scaffoldDirs, others...)
		}
	}
	return scaffoldDirs, nil
}
