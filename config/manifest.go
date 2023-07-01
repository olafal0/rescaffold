package config

import (
	"fmt"
	"io"
	"strings"

	"github.com/BurntSushi/toml"
)

type Manifest struct {
	Meta struct {
		Title       string `toml:"title"`
		Author      string `toml:"author"`
		Description string `toml:"description"`
	} `toml:"meta"`

	Config struct {
		OpenDelim     string `toml:"open_delim"`
		CloseDelim    string `toml:"close_delim"`
		ModifierDelim string `toml:"modifier_delim"`
	} `toml:"config"`

	Vars map[string]struct {
		Type        string   `toml:"type"`
		Description string   `toml:"description"`
		EnumValues  []string `toml:"enum_values"`
		Default     string   `toml:"default"`
	} `toml:"vars"`
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
