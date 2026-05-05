package yamlcfg_test

import (
	"embed"
	"fmt"
	"log/slog"

	"github.com/aranw/yamlcfg"
)

type Config struct {
	LogLevel string `yaml:"log_level"`
}

func ExampleParse() {
	cfg, err := yamlcfg.Parse[Config]("config.yaml")
	if err != nil {
		slog.Error("loading yaml config", "err", err)
		return
	}

	_ = cfg.LogLevel
}

//go:embed testdata
var testdata embed.FS

func ExampleParseFS() {
	cfg, err := yamlcfg.ParseFS[Config](testdata, "config.yaml")
	if err != nil {
		slog.Error("loading yaml config", "err", err)
		return
	}

	_ = cfg.LogLevel
}

func ExampleParseWithConfig() {
	defaults := &Config{
		LogLevel: "info",
	}

	cfg, err := yamlcfg.ParseWithConfig(defaults, "config.yaml")
	if err != nil {
		slog.Error("loading yaml config", "err", err)
		return
	}

	_ = cfg.LogLevel
}

func ExampleUnmarshalConfig() {
	data := []byte(`log_level: "debug"`)

	var cfg Config
	if err := yamlcfg.UnmarshalConfig(&cfg, data); err != nil {
		slog.Error("unmarshalling config", "err", err)
		return
	}

	fmt.Println(cfg.LogLevel)
	// Output: debug
}
