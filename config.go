package yamlcfg

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"regexp"
)

var envVarRe = regexp.MustCompile(`\$\{(\w+)(?::([^}]*))?\}`)

// ParseFS loads and unmarshals the config at the given path from an
// embed.FS. The config is also validated if it implements a Validate
// function. Decoding behaviour can be adjusted with opts (see Option).
func ParseFS[T any](fs embed.FS, path string, opts ...Option) (*T, error) {
	var cfg *T

	b, err := fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config from embed.FS: %w", err)
	}

	return parse(cfg, b, opts...)
}

// Parse reads the given path and unmarshals it into the given config
// type. The config is also validated if it has a Validate function on it.
// Decoding behaviour can be adjusted with opts (see Option).
func Parse[T any](path string, opts ...Option) (*T, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	var cfg *T
	return parse(cfg, b, opts...)
}

// ParseWithConfig parses the YAML file at the given path into the provided
// config, which can be pre-populated with default values; fields absent
// from the file retain their defaults.
// If the given config implements a `Validate() error` function this will be called.
// Decoding behaviour can be adjusted with opts (see Option).
func ParseWithConfig[T any](cfg *T, path string, opts ...Option) (*T, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	return parse(cfg, b, opts...)
}

// UnmarshalConfig takes the provided yaml data and unmarshals it
// into the provided config struct. It will return an error if the
// decoding fails or if the yaml data is not in the expected format.
// It will also expand any environment variables in the yaml data.
// Decoding behaviour can be adjusted with opts (see Option).
func UnmarshalConfig[T any](cfg *T, data []byte, opts ...Option) error {
	// expand any $VAR values in the config from environment variables
	data = []byte(parseEnv(string(data)))

	dec := newDecodeConfig(opts).decoder(bytes.NewReader(data))
	return dec.Decode(cfg)
}

func parse[T any](cfg *T, b []byte, opts ...Option) (*T, error) {
	if cfg == nil {
		cfg = new(T)
	}
	if err := UnmarshalConfig(cfg, b, opts...); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}
	if cfg, ok := any(cfg).(interface {
		Validate() error
	}); ok {
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("validating config: %w", err)
		}
	}

	return cfg, nil
}

// parseEnv replaces environment variables with their values.
// Supports two formats:
// 1. ${ENV_NAME} or ${ENV_NAME:default} anywhere in the string
// 2. $ENV_NAME only when it's the entire string
func parseEnv(input string) string {
	return envVarRe.ReplaceAllStringFunc(input, func(match string) string {
		parts := envVarRe.FindStringSubmatch(match)
		if len(parts) == 0 {
			return match
		}

		key := parts[1]
		defaultValue := parts[2]

		if value, found := os.LookupEnv(key); found {
			return value
		}
		return defaultValue
	})
}
