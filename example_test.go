package yamlcfg_test

import (
	"log/slog"

	"github.com/aranw/yamlcfg"
)

type Config struct {
	LogLevel string `yaml:"log_level"`
}

func ExampleLoad() {
	cfg, err := yamlcfg.Load[Config]("config.yaml")
	if err != nil {
		slog.Error("loading yaml config", "err", err)
		return
	}

	_ = cfg.LogLevel
}
