package scaffold

import (
	"bufio"
	"fmt"
	"os"

	"github.com/olafal0/rescaffold/config"
)

// LoadVarsInteractive loads vars from lockfile, and prompts the user for any
// vars that are in the manifest but not in the lockfile.
func LoadVarsInteractive(manifestVars map[string]*config.ManifestVar, lockfileVars map[string]string) (map[string]string, error) {
	varValues := make(map[string]string, len(manifestVars))
	for varName, varOptions := range manifestVars {
		if _, ok := lockfileVars[varName]; ok {
			varValues[varName] = lockfileVars[varName]
			continue
		}
		fmt.Printf("%s: %s\n", varName, varOptions.Description)
		if varOptions.Default != "" {
			fmt.Printf("Enter value [%s]: ", varOptions.Default)
		} else {
			fmt.Println("Enter value: ")
		}
		varValue := ""

		stdinScanner := bufio.NewScanner(os.Stdin)
		stdinScanner.Scan()
		varValue = stdinScanner.Text()
		if varValue == "" && varOptions.Default != "" {
			varValue = varOptions.Default
		}
		if varValue == "" && varOptions.Default == "" {
			return nil, fmt.Errorf("var %s is required", varName)
		}
		varValues[varName] = varValue
	}
	return varValues, nil
}
