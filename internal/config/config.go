package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
)

const Template = `[discord]
token = ""

[status]
prefix = "🎵 "
max_length = 70
clear_on_pause = true
clear_on_exit = true

[lyrics]
provider = "lrclib"
offset_ms = 500

[polling]
interval_ms = 300

[logging]
level = "info"
`

type Config struct {
	Discord Discord `toml:"discord"`
	Status  Status  `toml:"status"`
	Lyrics  Lyrics  `toml:"lyrics"`
	Polling Polling `toml:"polling"`
	Logging Logging `toml:"logging"`
}

type Discord struct {
	Token string `toml:"token"`
}

type Status struct {
	Prefix       string `toml:"prefix"`
	MaxLength    int    `toml:"max_length"`
	ClearOnPause bool   `toml:"clear_on_pause"`
	ClearOnExit  bool   `toml:"clear_on_exit"`
}

type Lyrics struct {
	Provider string `toml:"provider"`
	OffsetMS int    `toml:"offset_ms"`
}

type Polling struct {
	IntervalMS int `toml:"interval_ms"`
}

type Logging struct {
	Level string `toml:"level"`
}

func Default() Config {
	return Config{
		Status: Status{
			Prefix:       "🎵 ",
			MaxLength:    70,
			ClearOnPause: true,
			ClearOnExit:  true,
		},
		Lyrics: Lyrics{
			Provider: "lrclib",
			OffsetMS: 500,
		},
		Polling: Polling{
			IntervalMS: 300,
		},
		Logging: Logging{
			Level: "info",
		},
	}
}

func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	return ResolvePath(runtime.GOOS, os.Getenv, home)
}

func ResolvePath(goos string, getenv func(string) string, home string) (string, error) {
	switch goos {
	case "linux":
		if home == "" {
			return "", errors.New("cannot resolve Linux config path: home directory is unavailable")
		}
		return filepath.Join(home, ".config", "disgord-lyrics", "config.toml"), nil
	case "windows":
		base := getenv("ProgramData")
		if base == "" {
			base = getenv("APPDATA")
		}
		if base == "" {
			return "", errors.New("cannot resolve Windows config path: ProgramData and APPDATA are unavailable")
		}
		return filepath.Join(base, "DisGOrd Lyrics", "config.toml"), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", goos)
	}
}

func Init(path string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config file already exists: %s", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("check config file: %w", err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if !force {
		flags |= os.O_EXCL
	}

	file, err := os.OpenFile(path, flags, 0o600)
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return fmt.Errorf("secure config file: %w", err)
	}

	if _, err := file.WriteString(Template); err != nil {
		file.Close()
		return fmt.Errorf("write config file: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close config file: %w", err)
	}

	return nil
}

func Load(path string) (Config, error) {
	cfg := Default()

	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("load config file: %w", err)
	}
	if _, err := toml.Decode(string(content), &cfg); err != nil {
		return Config{}, errors.New("load config: invalid TOML")
	}

	cfg.Discord.Token = strings.TrimSpace(cfg.Discord.Token)
	cfg.Lyrics.Provider = strings.ToLower(strings.TrimSpace(cfg.Lyrics.Provider))
	cfg.Logging.Level = strings.ToLower(strings.TrimSpace(cfg.Logging.Level))

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	if strings.TrimSpace(cfg.Discord.Token) == "" {
		return errors.New("invalid config: discord token is required")
	}
	if cfg.Status.MaxLength < 1 || cfg.Status.MaxLength > 128 {
		return errors.New("invalid config: status max_length must be between 1 and 128")
	}
	if cfg.Lyrics.Provider != "lrclib" {
		return fmt.Errorf("invalid config: unsupported lyrics provider %q", cfg.Lyrics.Provider)
	}
	if cfg.Lyrics.OffsetMS < -60000 || cfg.Lyrics.OffsetMS > 60000 {
		return errors.New("invalid config: lyrics offset_ms must be between -60000 and 60000")
	}
	if cfg.Polling.IntervalMS < 100 || cfg.Polling.IntervalMS > 60000 {
		return errors.New("invalid config: polling interval_ms must be between 100 and 60000")
	}

	switch cfg.Logging.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid config: unsupported logging level %q", cfg.Logging.Level)
	}

	return nil
}
