package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Status.Prefix != "🎵 " || cfg.Status.MaxLength != 70 {
		t.Fatalf("unexpected status defaults: %#v", cfg.Status)
	}
	if cfg.Lyrics.Provider != "lrclib" || cfg.Lyrics.OffsetMS != 500 {
		t.Fatalf("unexpected lyrics defaults: %#v", cfg.Lyrics)
	}
	if cfg.Polling.IntervalMS != 300 || cfg.Logging.Level != "info" {
		t.Fatalf("unexpected runtime defaults: %#v %#v", cfg.Polling, cfg.Logging)
	}
}

func TestTemplateMatchesExample(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "config-example.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != Template {
		t.Fatal("config template and config-example.toml differ")
	}
}

func TestResolveLinuxPath(t *testing.T) {
	path, err := ResolvePath("linux", func(string) string { return "" }, "/home/tester")
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join("/home/tester", ".config", "disgord-lyrics", "config.toml")
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}

func TestResolveWindowsPathAndFallback(t *testing.T) {
	values := map[string]string{
		"ProgramData": `C:\ProgramData`,
		"APPDATA":     `C:\Users\tester\AppData\Roaming`,
	}
	getenv := func(key string) string { return values[key] }

	path, err := ResolvePath("windows", getenv, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(path, values["ProgramData"]) {
		t.Fatalf("expected ProgramData path, got %q", path)
	}

	values["ProgramData"] = ""
	path, err = ResolvePath("windows", getenv, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(path, values["APPDATA"]) {
		t.Fatalf("expected APPDATA fallback, got %q", path)
	}
}

func TestInitDoesNotOverwriteWithoutForce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "config.toml")

	if err := Init(path, false); err != nil {
		t.Fatal(err)
	}
	if err := Init(path, false); err == nil {
		t.Fatal("expected existing file error")
	}

	if err := os.WriteFile(path, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Init(path, true); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != Template {
		t.Fatal("expected forced init to restore template")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected config permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestLoadAppliesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("[discord]\ntoken = \"secret\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Status.MaxLength != 70 || cfg.Lyrics.Provider != "lrclib" || cfg.Polling.IntervalMS != 300 {
		t.Fatalf("defaults were not applied: %#v", cfg)
	}
}

func TestValidation(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected missing token error")
	}

	cfg.Discord.Token = "secret"
	cfg.Status.MaxLength = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected max length error")
	}

	cfg = Default()
	cfg.Discord.Token = "secret"
	cfg.Lyrics.Provider = "python-provider"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected provider error")
	}
}

func TestLoadDoesNotExposeMalformedToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	secret := "sensitive-token"
	content := "[discord]\ntoken = \"" + secret + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected TOML error")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("error exposed token")
	}
}
