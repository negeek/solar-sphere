// Package env loads KEY=VALUE pairs from a .env file into the process
// environment via os.Setenv, without pulling in a third-party dependency.
package env

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Load reads the .env file at path and applies each entry with os.Setenv.
// Variables already set in the environment are left untouched, so real
// environment variables always take precedence over the file. Blank lines
// and lines starting with '#' are ignored.
func Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for lineNum := 1; scanner.Scan(); lineNum++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("env: %s:%d: expected KEY=VALUE, got %q", path, lineNum, line)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("env: %s:%d: empty key", path, lineNum)
		}

		if _, isSet := os.LookupEnv(key); isSet {
			continue
		}
		if err := os.Setenv(key, unquote(strings.TrimSpace(value))); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func unquote(v string) string {
	if len(v) < 2 {
		return v
	}
	first, last := v[0], v[len(v)-1]
	if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
		return v[1 : len(v)-1]
	}
	return v
}
