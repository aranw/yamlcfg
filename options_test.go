package yamlcfg

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-quicktest/qt"
	"github.com/goccy/go-yaml"
)

type RefConfig struct {
	LogLevel string `yaml:"log_level"`
	Port     int    `yaml:"port"`
}

type failingValidator struct{}

func (failingValidator) Struct(v interface{}) error {
	return fmt.Errorf("validation failed for %T", v)
}

type passingValidator struct{}

func (passingValidator) Struct(v interface{}) error {
	return nil
}

func TestOptions(t *testing.T) {
	t.Run("AllowUnknownFields ignores unknown keys", func(t *testing.T) {
		b := []byte("some_value: \"foo\"\nunknown_value: \"bar\"\n")
		cfg, err := ParseWithConfig(&TestStruct{SecondValue: "default"}, "testdata/this_file_does_not_exist.yaml")
		qt.Assert(t, qt.IsNotNil(err)) // sanity: file doesn't exist

		cfg = &TestStruct{}
		err = UnmarshalConfig(cfg, b, AllowUnknownFields())
		qt.Assert(t, qt.IsNil(err))
		qt.Assert(t, qt.Equals(cfg.SomeValue, "foo"))
	})

	t.Run("unknown keys still rejected by default", func(t *testing.T) {
		b := []byte("some_value: \"foo\"\nunknown_value: \"bar\"\n")
		cfg := &TestStruct{}
		qt.Assert(t, qt.IsNotNil(UnmarshalConfig(cfg, b)))
	})

	t.Run("AllowDuplicateKeys keeps last value", func(t *testing.T) {
		b := []byte("some_value: \"first\"\nsome_value: \"second\"\n")
		cfg := &TestStruct{}
		qt.Assert(t, qt.IsNil(UnmarshalConfig(cfg, b, AllowDuplicateKeys())))
		qt.Assert(t, qt.Equals(cfg.SomeValue, "second"))
	})

	t.Run("duplicate keys rejected by default", func(t *testing.T) {
		b := []byte("some_value: \"first\"\nsome_value: \"second\"\n")
		cfg := &TestStruct{}
		qt.Assert(t, qt.IsNotNil(UnmarshalConfig(cfg, b)))
	})

	t.Run("AllowFieldPrefixes permits matching prefixes", func(t *testing.T) {
		b := []byte("some_value: \"foo\"\nsome_value_extra: \"bar\"\n")
		cfg := &TestStruct{}
		qt.Assert(t, qt.IsNil(UnmarshalConfig(cfg, b, AllowFieldPrefixes("some_value_"))))
		qt.Assert(t, qt.Equals(cfg.SomeValue, "foo"))
	})

	t.Run("AllowFieldPrefixes still rejects other unknown keys", func(t *testing.T) {
		b := []byte("some_value: \"foo\"\nother: \"bar\"\n")
		cfg := &TestStruct{}
		qt.Assert(t, qt.IsNotNil(UnmarshalConfig(cfg, b, AllowFieldPrefixes("some_value_"))))
	})

	t.Run("UseOrderedMap preserves key order", func(t *testing.T) {
		b := []byte("data:\n  z: 1\n  a: 2\n")
		cfg := struct {
			Data interface{} `yaml:"data"`
		}{}
		qt.Assert(t, qt.IsNil(UnmarshalConfig(&cfg, b, UseOrderedMap())))
		ms, ok := cfg.Data.(yaml.MapSlice)
		qt.Assert(t, qt.IsTrue(ok))
		qt.Assert(t, qt.Equals(len(ms), 2))
		qt.Assert(t, qt.Equals(fmt.Sprint(ms[0].Key), "z"))
		qt.Assert(t, qt.Equals(fmt.Sprint(ms[1].Key), "a"))
	})

	t.Run("WithValidator reports validation failures", func(t *testing.T) {
		b := []byte("log_level: \"info\"\nport: 8080\n")
		cfg := &RefConfig{}
		err := UnmarshalConfig(cfg, b, WithValidator(failingValidator{}))
		qt.Assert(t, qt.ErrorMatches(err, "validation failed for yamlcfg.RefConfig"))
	})

	t.Run("WithValidator accepts valid values", func(t *testing.T) {
		b := []byte("log_level: \"info\"\nport: 8080\n")
		cfg := &RefConfig{}
		qt.Assert(t, qt.IsNil(UnmarshalConfig(cfg, b, WithValidator(passingValidator{}))))
		qt.Assert(t, qt.Equals(cfg.Port, 8080))
	})

	t.Run("WithReferenceFiles resolves aliases", func(t *testing.T) {
		b := []byte("log_level: *log_level\nport: *port\n")
		cfg := &RefConfig{}
		qt.Assert(t, qt.IsNil(UnmarshalConfig(cfg, b, WithReferenceFiles("testdata/anchor.yaml"))))
		qt.Assert(t, qt.Equals(cfg.LogLevel, "info"))
		qt.Assert(t, qt.Equals(cfg.Port, 8080))
	})

	t.Run("WithReferenceDirs resolves aliases", func(t *testing.T) {
		cfg, err := Parse[RefConfig]("testdata/refs/config.yaml", WithReferenceDirs("testdata/refs"))
		qt.Assert(t, qt.IsNil(err))
		qt.Assert(t, qt.Equals(cfg.LogLevel, "debug"))
		qt.Assert(t, qt.Equals(cfg.Port, 9090))
	})

	t.Run("WithRecursiveDirs resolves aliases", func(t *testing.T) {
		cfg, err := Parse[RefConfig]("testdata/refs/config.yaml", WithRecursiveDirs("testdata"))
		qt.Assert(t, qt.IsNil(err))
		qt.Assert(t, qt.Equals(cfg.LogLevel, "debug"))
		qt.Assert(t, qt.Equals(cfg.Port, 9090))
	})

	t.Run("WithEmbeddedReferences resolves aliases from embed.FS", func(t *testing.T) {
		cfg, err := ParseFS[RefConfig](testdata, "testdata/refs/config.yaml",
			WithEmbeddedReferences(testdata, "testdata/refs/anchor.yaml"))
		qt.Assert(t, qt.IsNil(err))
		qt.Assert(t, qt.Equals(cfg.LogLevel, "debug"))
		qt.Assert(t, qt.Equals(cfg.Port, 9090))
	})

	t.Run("WithEmbeddedReferences errors on missing file", func(t *testing.T) {
		b := []byte("log_level: *log_level\n")
		cfg := &RefConfig{}
		qt.Assert(t, qt.IsNotNil(UnmarshalConfig(cfg, b, WithEmbeddedReferences(testdata, "testdata/missing.yaml"))))
	})

	t.Run("references without option fail", func(t *testing.T) {
		b := []byte("log_level: *log_level\n")
		cfg := &RefConfig{}
		qt.Assert(t, qt.IsNotNil(UnmarshalConfig(cfg, b)))
	})

	t.Run("options compose", func(t *testing.T) {
		b := []byte("log_level: *log_level\nport: *port\nextra: \"ignored\"\n")
		cfg := &RefConfig{}
		qt.Assert(t, qt.IsNil(UnmarshalConfig(cfg, b,
			WithReferenceFiles("testdata/anchor.yaml"),
			AllowUnknownFields(),
			WithValidator(passingValidator{}),
		)))
		qt.Assert(t, qt.Equals(cfg.LogLevel, "info"))
	})
}

func TestOptionErrorsWrapTypedErrors(t *testing.T) {
	// errors from Parse (which wrap with a prefix) still expose typed errors
	_, err := Parse[RefConfig]("testdata/refs/config.yaml")
	qt.Assert(t, qt.IsNotNil(err))
	qt.Assert(t, qt.IsTrue(IsSyntaxError(errors.Unwrap(err)) || IsSyntaxError(err)))
}
