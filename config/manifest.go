package config

import (
	"fmt"
	"io"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	ManifestFilename = ".rescaffold-manifest.toml"
)

type Manifest struct {
	Meta *ManifestMeta `toml:"meta"`

	Config *ManifestConfig `toml:"config"`

	Vars map[string]*ManifestVar `toml:"vars"`
}

type ManifestMeta struct {
	Title       string `toml:"title"`
	Author      string `toml:"author"`
	Description string `toml:"description"`
}

type ManifestVar struct {
	Type        string   `toml:"type"`
	Description string   `toml:"description"`
	EnumValues  []string `toml:"enum_values"`
	Default     string   `toml:"default"`
}

type ManifestConfig struct {
	OpenDelim     string `toml:"open_delim"`
	CloseDelim    string `toml:"close_delim"`
	ModifierDelim string `toml:"modifier_delim"`
}

func ParseManifest(data io.Reader) (*Manifest, error) {
	manifest := &Manifest{}
	meta, err := toml.NewDecoder(data).Decode(manifest)
	if err != nil {
		return nil, err
	}

	undecodedKeys := meta.Undecoded()
	if len(undecodedKeys) > 0 {
		return nil, fmt.Errorf("unknown keys in manifest: %v", undecodedKeys)
	}
	return manifest, nil
}

func (m *Manifest) String() string {
	buf := &strings.Builder{}
	if err := toml.NewEncoder(buf).Encode(m); err != nil {
		return fmt.Sprintf("unencodable: %v", err)
	}
	return buf.String()
}
