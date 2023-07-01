package config_test

import (
	"bytes"
	"testing"

	"github.com/olafal0/rescaffold/assert"
	"github.com/olafal0/rescaffold/config"
)

func TestParseManifest(t *testing.T) {
	data := `
[meta]
title = "Example Scaffold"
author = "me"

[config]
open_delim = "_"
close_delim = "_"

[vars.project_name]
type = "string"
description = "A short, descriptive name for your project"
`
	manifest, err := config.ParseManifest(bytes.NewBuffer([]byte(data)))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, manifest.Meta.Title, "Example Scaffold")
	assert.Equal(t, manifest.Config.OpenDelim, "_")
	assert.Equal(t, manifest.Vars["project_name"].Type, "string")
	assert.StrNotContains(t, manifest.String(), "unencodable")
}

func TestParseUnexpectedFields(t *testing.T) {
	data := `
[meta]
title = "Example Scaffold"
author = "me"
author_email = "me@example.com"
`
	_, err := config.ParseManifest(bytes.NewBuffer([]byte(data)))
	if err == nil {
		t.Error("expected error")
	}
	t.Log(err)

	data = `this isn't toml at all`
	_, err = config.ParseManifest(bytes.NewBuffer([]byte(data)))
	if err == nil {
		t.Error("expected error")
	}
	t.Log(err)
}
