package yamlcfg

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// ParseFS attempts to load the given path and config from a embed.FS
func ParseFS[T any](fs embed.FS, path string) (*T, error) {
	var cfg *T

	b, err := fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config from embed.FS: %w", err)
	}

	return parse(cfg, b)
}

// Parse takes the given path and attempts to read and unmarshal the config
// The given config will also be validated if it has a Validate function on it
func Parse[T any](path string) (*T, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	var cfg *T
	return parse(cfg, b)
}

// ParseWithConfig takes the given config and path and attempts to read and unmarshal the config
// The given config can be populated with default values
// If the given config implements a `Validate() error` function this will be called
func ParseWithConfig[T any](cfg *T, path string) (*T, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	return parse(cfg, b)
}

// UnmarshalConfig takes the provided yaml data and unmarshals it
// into the provided config struct. It will return an error if the
// decoding fails or if the yaml data is not in the expected format.
// It will also expand any environment variables in the yaml data
func UnmarshalConfig[T any](cfg *T, data []byte) error {
	// expand any $VAR values in the config from environment variables
	data = []byte(parseEnv(string(data)))

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	return dec.Decode(&cfg)
}

func parse[T any](cfg *T, b []byte) (*T, error) {
	if err := UnmarshalConfig(&cfg, b); err != nil {
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

// parseEnv replaces environment variables with their values.
// Supports two formats:
// 1. ${ENV_NAME} or ${ENV_NAME:default} anywhere in the string
// 2. $ENV_NAME only when it's the entire string
func parseEnv(input string) string {
	re := regexp.MustCompile(`\$\{(\w+)(?::([^}]*))?\}`)

	return re.ReplaceAllStringFunc(input, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 0 {
			return match
		}

		key := parts[1]
		defaultValue := parts[2] // May be empty if no default provided

		if value, found := os.LookupEnv(key); found {
			return value
		}
		return defaultValue // Return default value (empty string if no default was provided)
	})
}
