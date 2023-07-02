package scaffold

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/olafal0/rescaffold/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var Modifiers = map[string]func(string) string{
	"titlecase": cases.Title(language.English).String,
	"lowercase": strings.ToLower,
	"uppercase": strings.ToUpper,
}

// LiteralMatchReplacer returns a function that will perform template replacement on a string
// or substring, using the configured delimiters and vars. This function supports
// zero or one modifiers per variable.
func LiteralMatchReplacer(manifest *config.Manifest, vars map[string]string) func(string) string {
	openDelim := manifest.Config.OpenDelim
	closeDelim := manifest.Config.CloseDelim
	modifierDelim := manifest.Config.ModifierDelim
	// Simplest implementation: create map of all vars * all modifiers and their
	// replacement value
	// TODO: support multiple modifiers
	replacements := make(map[string]string, len(vars))
	for varName, varValue := range vars {
		replacements[openDelim+varName+closeDelim] = varValue
		for modifierName, modifier := range Modifiers {
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

func RegexpReplacer(manifest *config.Manifest, vars map[string]string) (func(string) string, error) {
	openDelim := regexp.QuoteMeta(manifest.Config.OpenDelim)
	closeDelim := regexp.QuoteMeta(manifest.Config.CloseDelim)
	modifierDelim := regexp.QuoteMeta(manifest.Config.ModifierDelim)
	// Create regex from var and modifier names
	// Example: "x_(name|port)((?:\|(?:titlecase|lowercase))*)_"
	regexpStr := fmt.Sprintf(`%[1]s(%[2]s)((?:%[3]s(?:%[4]s))*)%[5]s`,
		openDelim,
		strings.Join(keys(vars), "|"),
		modifierDelim,
		strings.Join(keys(Modifiers), "|"),
		closeDelim,
	)
	matcher, err := regexp.Compile(regexpStr)
	if err != nil {
		return nil, err
	}

	return func(s string) string {
		submatches := matcher.FindAllStringSubmatchIndex(s, -1)
		if len(submatches) == 0 {
			return s
		}
		// Track replaced string differently from original so that we can always
		// read from the original, even after some replacements may have occurred.
		replacedStr := s
		// Since we're going left-to-right, track an offset that needs to be used
		// only when replacing segments of the modified string.
		replacementOffset := 0

		for _, submatch := range submatches {
			// Submatch indices are start and end index pairs.
			// First pair is the start and end of the entire match.
			// Second pair, in this case, should be the variable name.
			// Third pair, if it exists, should be the modifier chain (e.g. "|titlecase|lowercase")
			if len(submatch) < 4 {
				continue
			}
			entireStart := submatch[0]
			entireEnd := submatch[1]
			varNameStart := submatch[2]
			varNameEnd := submatch[3]
			varName := s[varNameStart:varNameEnd]
			varValue := vars[varName]

			// Look for modifiers if there is a third pair
			if len(submatch) >= 6 && submatch[4] != -1 && submatch[5] != -1 && submatch[5]-submatch[4] > 0 {
				modifierStart := submatch[4]
				modifierEnd := submatch[5]
				// Trim leading modifier delimiter
				modifierStr := strings.TrimPrefix(s[modifierStart:modifierEnd], manifest.Config.ModifierDelim)
				// Split modifier string on delimiter to create a left-to-right list of modifier names
				modifierNames := strings.Split(modifierStr, manifest.Config.ModifierDelim)
				for _, modifierName := range modifierNames {
					// We don't need to check for existence of modifier, because the regex
					// will only match on the modifier names
					varValue = Modifiers[modifierName](varValue)
				}
			}

			replacedStr = replacedStr[:entireStart-replacementOffset] + varValue + replacedStr[entireEnd-replacementOffset:]
			replacementOffset += (entireEnd - entireStart) - len(varValue)
		}
		return replacedStr
	}, nil
}

func RegexpLoopReplacer(manifest *config.Manifest, vars map[string]string) (func(string) string, error) {
	openDelim := regexp.QuoteMeta(manifest.Config.OpenDelim)
	closeDelim := regexp.QuoteMeta(manifest.Config.CloseDelim)
	modifierDelim := regexp.QuoteMeta(manifest.Config.ModifierDelim)
	// Create regex from var and modifier names
	// Example: "x_(name|port)((?:\|(?:titlecase|lowercase))*)_"
	regexpStr := fmt.Sprintf(`%[1]s(%[2]s)((?:%[3]s(?:%[4]s))*)%[5]s`,
		openDelim,
		strings.Join(keys(vars), "|"),
		modifierDelim,
		strings.Join(keys(Modifiers), "|"),
		closeDelim,
	)
	matcher, err := regexp.Compile(regexpStr)
	if err != nil {
		return nil, err
	}

	return func(s string) string {
		// Perform replacement one match at a time, left to right, to avoid complicated
		// logic of tracking offsets
		for {
			submatch := matcher.FindStringSubmatchIndex(s)
			if submatch == nil {
				break
			}
			// Submatch indices are start and end index pairs.
			// First pair is the start and end of the entire match.
			// Second pair, in this case, should be the variable name.
			// Third pair, if it exists, should be the modifier chain (e.g. "|titlecase|lowercase")
			if len(submatch) < 4 {
				continue
			}
			entireStart := submatch[0]
			entireEnd := submatch[1]
			varNameStart := submatch[2]
			varNameEnd := submatch[3]
			varName := s[varNameStart:varNameEnd]
			varValue := vars[varName]

			// Look for modifiers if there is a third pair
			if len(submatch) >= 6 && submatch[4] != -1 && submatch[5] != -1 && submatch[5]-submatch[4] > 0 {
				modifierStart := submatch[4]
				modifierEnd := submatch[5]
				// Trim leading modifier delimiter
				modifierStr := strings.TrimPrefix(s[modifierStart:modifierEnd], manifest.Config.ModifierDelim)
				// Split modifier string on delimiter to create a left-to-right list of modifier names
				modifierNames := strings.Split(modifierStr, manifest.Config.ModifierDelim)
				for _, modifierName := range modifierNames {
					// We don't need to check for existence of modifier, because the regex
					// will only match on the modifier names
					varValue = Modifiers[modifierName](varValue)
				}
			}

			s = s[:entireStart] + varValue + s[entireEnd:]
		}
		return s
	}, nil
}

func keys[T comparable, T2 any](m map[T]T2) []T {
	keys := make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
