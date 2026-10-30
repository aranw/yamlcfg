package yamlcfg_test

import (
	"embed"
	"errors"
	"fmt"

	"github.com/aranw/yamlcfg"
)

// Config is a minimal example config type.
type Config struct {
	LogLevel string `yaml:"log_level"`
}

// ServerConfig demonstrates a config with defaults and multiple fields.
type ServerConfig struct {
	LogLevel string `yaml:"log_level"`
	Port     int    `yaml:"port"`
}

//go:embed testdata/refs
var refsFS embed.FS

func ExampleParse() {
	cfg, err := yamlcfg.Parse[Config]("testdata/example.yaml")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cfg.LogLevel)
	// Output: info
}

func ExampleParseWithConfig() {
	// Fields absent from the YAML file keep their pre-populated defaults.
	defaults := &ServerConfig{
		LogLevel: "debug",
		Port:     8080,
	}

	cfg, err := yamlcfg.ParseWithConfig(defaults, "testdata/example.yaml")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cfg.LogLevel, cfg.Port)
	// Output: info 8080
}

func ExampleParseFS() {
	// config.yaml uses aliases defined in anchor.yaml, embedded alongside
	// it and resolved with WithEmbeddedReferences.
	cfg, err := yamlcfg.ParseFS[ServerConfig](refsFS, "testdata/refs/config.yaml",
		yamlcfg.WithEmbeddedReferences(refsFS, "testdata/refs/anchor.yaml"))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cfg.LogLevel, cfg.Port)
	// Output: debug 9090
}

func ExampleUnmarshalConfig() {
	data := []byte(`log_level: "debug"`)

	var cfg Config
	if err := yamlcfg.UnmarshalConfig(&cfg, data); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cfg.LogLevel)
	// Output: debug
}

func ExampleUnmarshalConfig_allowUnknownFields() {
	data := []byte("log_level: \"info\"\nunknown_key: \"ignored\"\n")

	var cfg Config
	if err := yamlcfg.UnmarshalConfig(&cfg, data, yamlcfg.AllowUnknownFields()); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cfg.LogLevel)
	// Output: info
}

func ExampleUnmarshalConfig_withReferences() {
	// testdata/refs/anchor.yaml defines:
	//
	//	log_level: &log_level debug
	//	port: &port 9090
	data := []byte("log_level: *log_level\nport: *port\n")

	var cfg ServerConfig
	err := yamlcfg.UnmarshalConfig(&cfg, data,
		yamlcfg.WithReferenceFiles("testdata/refs/anchor.yaml"))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cfg.LogLevel, cfg.Port)
	// Output: debug 9090
}

// portValidator implements yamlcfg.StructValidator.
type portValidator struct{}

func (portValidator) Struct(v any) error {
	cfg, ok := v.(ServerConfig)
	if !ok {
		return nil
	}
	if cfg.Port < 1 {
		return fmt.Errorf("port must be positive, got %d", cfg.Port)
	}
	return nil
}

func ExampleUnmarshalConfig_withValidator() {
	data := []byte("log_level: \"info\"\nport: 0\n")

	var cfg ServerConfig
	if err := yamlcfg.UnmarshalConfig(&cfg, data, yamlcfg.WithValidator(portValidator{})); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfg.LogLevel, cfg.Port)
	// Output: port must be positive, got 0
}

func ExampleUnmarshalConfig_typedErrors() {
	data := []byte("log_level: \"info\"\nunknown_key: 1\n")

	var cfg Config
	err := yamlcfg.UnmarshalConfig(&cfg, data)

	var fieldErr *yamlcfg.UnknownFieldError
	if errors.As(err, &fieldErr) {
		line, column, _ := yamlcfg.Position(err)
		fmt.Printf("unknown field at %d:%d\n", line, column)
	}
	// Output: unknown field at 2:1
}
