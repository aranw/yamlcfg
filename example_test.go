package yamlcfg_test

import (
	"embed"
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
