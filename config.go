package yamlcfg

import (
	"bytes"
	"embed"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseFS attempts to load the given path and config from a embed.FS
func ParseFS[T any](fs embed.FS, path string) (*T, error) {
	var cfg *T

	b, err := fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config from embed.FS: %w", err)
	} else if err := UnmarshalConfig(&cfg, b); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}
	if cfg, ok := interface{}(cfg).(interface {
		Validate() error
	}); ok {
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("validating config: %w", err)
		}
	}

	return cfg, nil
}

// Parse takes the given path and attempts to read and unmarshal the config
// The given config will also be validated if it has a Validate function on it
func Parse[T any](path string) (*T, error) {
	var cfg *T

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	} else if err := UnmarshalConfig(&cfg, b); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}
	if cfg, ok := interface{}(cfg).(interface {
		Validate() error
	}); ok {
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("validating config: %w", err)
		}
	}

	return cfg, nil
}

// UnmarshalConfig takes the provided yaml data and unmarshals it
// into the provided config struct. It will return an error if the
// decoding fails or if the yaml data is not in the expected format.
// It will also expand any environment variables in the yaml data
func UnmarshalConfig[T any](cfg *T, data []byte) error {
	// expand any $VAR values in the config from environment variables
	data = []byte(os.ExpandEnv(string(data)))

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	return dec.Decode(&cfg)
}
