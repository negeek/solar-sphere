package env

import (
	"os"
	"path/filepath"
	"testing"
)

func writeEnvFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write test .env file: %v", err)
	}
	return path
}

func TestLoadSetsVariables(t *testing.T) {
	path := writeEnvFile(t, "FOO=bar\nBAZ=qux\n")
	os.Unsetenv("FOO")
	os.Unsetenv("BAZ")

	if err := Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := os.Getenv("FOO"); got != "bar" {
		t.Errorf("FOO = %q, want bar", got)
	}
	if got := os.Getenv("BAZ"); got != "qux" {
		t.Errorf("BAZ = %q, want qux", got)
	}
}

func TestLoadSkipsBlankLinesAndComments(t *testing.T) {
	path := writeEnvFile(t, "\n# a comment\nFOO=bar\n   \n# another\n")
	os.Unsetenv("FOO")

	if err := Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := os.Getenv("FOO"); got != "bar" {
		t.Errorf("FOO = %q, want bar", got)
	}
}

func TestLoadStripsQuotes(t *testing.T) {
	path := writeEnvFile(t, `FOO="bar baz"`+"\n"+`SINGLE='qux'`+"\n")
	os.Unsetenv("FOO")
	os.Unsetenv("SINGLE")

	if err := Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := os.Getenv("FOO"); got != "bar baz" {
		t.Errorf("FOO = %q, want %q", got, "bar baz")
	}
	if got := os.Getenv("SINGLE"); got != "qux" {
		t.Errorf("SINGLE = %q, want qux", got)
	}
}

func TestLoadDoesNotOverrideExistingEnv(t *testing.T) {
	path := writeEnvFile(t, "FOO=from-file\n")
	t.Setenv("FOO", "from-environment")

	if err := Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := os.Getenv("FOO"); got != "from-environment" {
		t.Errorf("FOO = %q, want from-environment (real env must win)", got)
	}
}

func TestLoadRejectsMalformedLine(t *testing.T) {
	path := writeEnvFile(t, "this line has no equals sign\n")

	if err := Load(path); err == nil {
		t.Fatal("expected an error for a line without '='")
	}
}

func TestLoadMissingFile(t *testing.T) {
	if err := Load(filepath.Join(t.TempDir(), "does-not-exist.env")); err == nil {
		t.Fatal("expected an error for a missing file")
	}
}
