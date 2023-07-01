package scaffold

import (
	"strings"

	"github.com/olafal0/rescaffold/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var modifiers = map[string]func(string) string{
	"titlecase": cases.Title(language.English).String,
	"lowercase": strings.ToLower,
	"uppercase": strings.ToUpper,
}

// Replacer returns a function that will perform template replacement on a string
// or substring, using the configured delimiters and vars.
func Replacer(manifest *config.Manifest, vars map[string]string) func(string) string {
	openDelim := manifest.Config.OpenDelim
	closeDelim := manifest.Config.CloseDelim
	modifierDelim := manifest.Config.ModifierDelim
	// Simplest implementation: create map of all vars * all modifiers and their
	// replacement value
	// TODO: support multiple modifiers
	replacements := make(map[string]string, len(vars))
	for varName, varValue := range vars {
		replacements[openDelim+varName+closeDelim] = varValue
		for modifierName, modifier := range modifiers {
			replacements[openDelim+varName+modifierDelim+modifierName+closeDelim] = modifier(varValue)
		}
	}

	return func(s string) string {
		for k, v := range replacements {
			s = strings.ReplaceAll(s, k, v)
		}
		return s
	}
}
