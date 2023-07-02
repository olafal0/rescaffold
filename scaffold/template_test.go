package scaffold_test

import (
	"testing"

	"github.com/olafal0/rescaffold/assert"
	"github.com/olafal0/rescaffold/config"
	"github.com/olafal0/rescaffold/scaffold"
)

var (
	Vars = map[string]string{
		"name": "MyApp",
		"port": "8080",
	}
)

func testMakeManifest() *config.Manifest {
	manifest := &config.Manifest{
		Config: &config.ManifestConfig{
			OpenDelim:     "x_",
			CloseDelim:    "_",
			ModifierDelim: "|",
		},
		Vars: make(map[string]*config.ManifestVar, len(Vars)),
	}
	for k := range Vars {
		manifest.Vars[k] = &config.ManifestVar{
			Type: "string",
		}
	}
	return manifest
}

func testReplacer(t *testing.T, replacer func(string) string) {
	assert.Equal(t, replacer("foo"), "foo")
	assert.Equal(t, replacer("x_foo_"), "x_foo_")
	assert.Equal(t, replacer("x_names_"), "x_names_")
	assert.Equal(t, replacer("x_name_"), "MyApp")
	assert.Equal(t, replacer("x_name|lowercase_"), "myapp")
	assert.Equal(t, replacer("x_name|titlecase_"), "Myapp")
	assert.Equal(t, replacer("This is my app, x_name|uppercase|lowercase|titlecase_, running on port x_port_"), "This is my app, Myapp, running on port 8080")
}

func TestReplacers(t *testing.T) {
	manifest := testMakeManifest()
	replacer, err := scaffold.RegexpReplacer(manifest, Vars)
	if err != nil {
		t.Fatal(err)
	}

	testReplacer(t, replacer)

	replacer, err = scaffold.RegexpLoopReplacer(manifest, Vars)
	if err != nil {
		t.Fatal(err)
	}

	testReplacer(t, replacer)

	// Test simple replacer without the multiple modifier case
	replacer = scaffold.LiteralMatchReplacer(manifest, Vars)
	assert.Equal(t, replacer("foo"), "foo")
	assert.Equal(t, replacer("x_foo_"), "x_foo_")
	assert.Equal(t, replacer("x_names_"), "x_names_")
	assert.Equal(t, replacer("x_name_"), "MyApp")
	assert.Equal(t, replacer("x_name|lowercase_"), "myapp")
	assert.Equal(t, replacer("x_name|titlecase_"), "Myapp")
}

func BenchmarkRegexpReplacer(b *testing.B) {
	manifest := testMakeManifest()
	replacer, err := scaffold.RegexpReplacer(manifest, Vars)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		replacer("This is my app, x_name|uppercase|lowercase|titlecase_, running on port x_port_")
	}
}

func BenchmarkRegexpLoopReplacer(b *testing.B) {
	manifest := testMakeManifest()
	replacer, err := scaffold.RegexpLoopReplacer(manifest, Vars)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		replacer("This is my app, x_name|uppercase|lowercase|titlecase_, running on port x_port_")
		replacer("This is a string without any replacement")
	}
}
