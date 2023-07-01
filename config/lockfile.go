package config

import (
	"fmt"
	"io"

	"github.com/BurntSushi/toml"
)

type Lockfile struct {
}

func ParseLockfile(data io.Reader) (*Lockfile, error) {
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
