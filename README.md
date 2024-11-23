# yamlcfg

yamlcfg is a wrapper around the [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) library and provides a convient way to configure Golang applications with YAML and environment variables.

The library can also automatically call `Validate` functions if present on the given config struct.

## Installation

To install, run:

```
go get github.com/aranw/yamlcfg
```

## License

The yamlcfg package is licensed under the MIT. Please see the LICENSE file for details.

## Example

Example `config.yaml` with configuration value loaded from environment variables

```yaml
log_level: $LOG_LEVEL
```

Example `main.go`

```go
package main

import (
	"log/slog"

	"github.com/aranw/yamlcfg"
)

type Config struct {
	LogLevel string `yaml:"log_level"`
}

func (c *Config) Validate() error {
    validLevels := [...]string{"debug", "info", "error"}

	validLevel := slices.ContainsFunc(validLevels[:], func(s string) bool {
		return strings.EqualFold(s, c.Log.Level)
	})

	if c.Log.Level == "" {
		c.Log.Level = "info"
	} else if !validLevel {
		return errors.New("invalid log level provided")
	}

    return nil
}

func main() {
	cfg, err := yamlcfg.Parse[Config]("config.yaml")
	if err != nil {
		slog.Error("parsing yaml config", "err", err)
		return
	}

	_ = cfg.LogLevel
}

```

## Example with default value


Example `config.yaml` with environment configuration and default values

```yaml
log_level: ${LOG_LEVEL:info}
```
