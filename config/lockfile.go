package config

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

const (
	LockfileFilename = ".rescaffold.toml"
)

type Lockfile struct {
	Scaffolds map[string]*LockfileScaffold `toml:"scaffolds"`

	// filename is the filename from which this lockfile was loaded or created
	filename     string
	newlyCreated bool
}

type LockfileScaffold struct {
	Source string                  `toml:"source"`
	Files  []*LockfileScaffoldFile `toml:"file"`
	Vars   map[string]string       `toml:"vars"`
}

type LockfileScaffoldFile struct {
	Path     string `toml:"path"`
	Checksum string `toml:"checksum"`
}

// LoadLockfile loads a lockfile from the given filename. If the file does not
// exist, a new lockfile is created and returned.
//
// If a lockfile is returned, it will have an associated filename and can be
// written back to disk.
func LoadLockfile(filename string) (*Lockfile, error) {
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return CreateLockfile(filename)
		}
		return nil, fmt.Errorf("opening lockfile for reading failed: %w", err)
	}
	lockfile, err := parseLockfile(f)
	if err != nil {
		return nil, fmt.Errorf("parse lockfile failed: %w", err)
	}
	lockfile.filename = filename
	return lockfile, nil
}

// CreateLockfile creates a new default lockfile and writes it to the given
// filename.
//
// If a lockfile is returned, it will have an associated filename and can be
// written back to disk.
func CreateLockfile(filename string) (*Lockfile, error) {
	if err := os.MkdirAll(path.Dir(filename), 0755); err != nil {
		return nil, err
	}
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	lockfile := defaultLockfile()
	if err := toml.NewEncoder(f).Encode(lockfile); err != nil {
		return nil, err
	}

	lockfile.filename = filename
	lockfile.newlyCreated = true
	log.Printf("lockfile created: %s\n", filename)
	return lockfile, nil
}

// WriteUpdated writes the lockfile back to disk in its current state
func (l *Lockfile) WriteUpdated() error {
	if l.filename == "" {
		return fmt.Errorf("lockfile has unknown filename")
	}
	// Write lockfile to buffer so that, in the event of an encoding error, the
	// lockfile is not truncated
	buf := &bytes.Buffer{}
	enc := toml.NewEncoder(buf)
	enc.Indent = ""
	err := enc.Encode(l)
	if err != nil {
		return fmt.Errorf("failed to encode lockfile for update: %w", err)
	}

	f, err := os.Create(l.filename)
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(f)
	if err != nil {
		return err
	}
	return nil
}

func (l *Lockfile) IsNewlyCreated() bool {
	return l.newlyCreated
}

// Remove removes the lockfile from disk permanently
func (l *Lockfile) Remove() error {
	return os.Remove(l.filename)
}

// GetScaffold returns the lockfile information for a scaffold, initializing it
// if it doesn't exist
func (l *Lockfile) GetScaffold(name, source string, manifest *Manifest) *LockfileScaffold {
	ls, ok := l.Scaffolds[name]
	if ok {
		return ls
	}

	newLS := &LockfileScaffold{
		Source: source,
		Files:  []*LockfileScaffoldFile{},
		Vars:   map[string]string{},
	}
	l.Scaffolds[name] = newLS
	return newLS
}

// GetFile returns the lockfile information for a file. If the file does not
// exist, it returns nil.
func (ls *LockfileScaffold) GetFile(path string) *LockfileScaffoldFile {
	for _, f := range ls.Files {
		if f.Path == path {
			return f
		}
	}
	return nil
}

// SetFile sets the lockfile information for a file, creating it if it doesn't exist
func (ls *LockfileScaffold) SetFile(path, checksum string) *LockfileScaffoldFile {
	lockedFile := ls.GetFile(path)
	if lockedFile != nil {
		lockedFile.Checksum = checksum
		return lockedFile
	}

	lockedFile = &LockfileScaffoldFile{
		Path:     path,
		Checksum: checksum,
	}
	ls.Files = append(ls.Files, lockedFile)
	return lockedFile
}

func (ls *LockfileScaffold) RemoveFile(path string) {
	for i, f := range ls.Files {
		if f.Path == path {
			ls.Files = append(ls.Files[:i], ls.Files[i+1:]...)
			return
		}
	}
}

func defaultLockfile() *Lockfile {
	return &Lockfile{
		Scaffolds: map[string]*LockfileScaffold{},
	}
}

func parseLockfile(data io.Reader) (*Lockfile, error) {
	lockfile := &Lockfile{}
	meta, err := toml.NewDecoder(data).Decode(lockfile)
	if err != nil {
		return nil, err
	}
	undecodedKeys := meta.Undecoded()
	if len(undecodedKeys) > 0 {
		return nil, fmt.Errorf("unknown keys in lockfile: %v", undecodedKeys)
	}

	return lockfile, nil
}
