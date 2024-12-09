package yamlcfg

import (
	"embed"
	"fmt"
	"testing"

	"github.com/go-quicktest/qt"
)

type TestStruct struct {
	SomeValue   string `yaml:"some_value"`
	SecondValue string `yaml:"second_value"`
}

func (t *TestStruct) Validate() error {
	return nil
}

type TestStructWithFailingValidation struct {
	SomeValue string `yaml:"some_value"`
}

func (t *TestStructWithFailingValidation) Validate() error {
	return fmt.Errorf("this is going to fail")
}

//go:embed testdata/*
var testdata embed.FS

func TestParseFS(t *testing.T) {

	t.Run("successfully load and unmarshals config", func(t *testing.T) {
		cfg, err := ParseFS[TestStruct](testdata, "testdata/test1.yaml")
		if err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.SomeValue, "this is for testing purposes"))
	})

	t.Run("fails to read unknown file", func(t *testing.T) {
		cfg, err := ParseFS[TestStruct](testdata, "testdata/this_file_does_not_exist.yaml")
		qt.Assert(t, qt.IsNotNil(err))
		qt.Assert(t, qt.ErrorMatches(err, "reading config from embed.FS: .*"))
		qt.Assert(t, qt.IsNil(cfg))
	})

	t.Run("fails to read wrong file type", func(t *testing.T) {
		cfg, err := ParseFS[TestStruct](testdata, "testdata/gopher.png")
		qt.Assert(t, qt.IsNotNil(err))
		qt.Assert(t, qt.ErrorMatches(err, "unmarshalling config: yaml: .*"))
		qt.Assert(t, qt.IsNil(cfg))
	})

	t.Run("fails to validate config struct", func(t *testing.T) {
		cfg, err := ParseFS[TestStructWithFailingValidation](testdata, "testdata/test1.yaml")
		qt.Assert(t, qt.IsNotNil(err))
		qt.Assert(t, qt.ErrorMatches(err, "validating config: this is going to fail"))
		qt.Assert(t, qt.IsNil(cfg))
	})
}

func TestParseWithConfig(t *testing.T) {
	t.Run("successfully load and unmarshals config with defaults", func(t *testing.T) {
		cfg := &TestStruct{
			SecondValue: "This is the default",
		}

		cfg, err := ParseWithConfig[TestStruct](cfg, "testdata/test1.yaml")
		if err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.SomeValue, "this is for testing purposes"))
		qt.Assert(t, qt.Equals(cfg.SecondValue, "This is the default"))
	})

}

func TestParse(t *testing.T) {

	t.Run("successfully load and unmarshals config", func(t *testing.T) {
		cfg, err := Parse[TestStruct]("testdata/test1.yaml")
		if err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.SomeValue, "this is for testing purposes"))
	})

	t.Run("successfully parses yaml config with default value", func(t *testing.T) {
		cfg, err := Parse[TestStruct]("testdata/test_with_default.yaml")
		if err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.SomeValue, "this_is_the_default"))
	})

	t.Run("fails to read unknown file", func(t *testing.T) {
		cfg, err := Parse[TestStruct]("testdata/this_file_does_not_exist.yaml")
		qt.Assert(t, qt.IsNotNil(err))
		qt.Assert(t, qt.ErrorMatches(err, "reading config file: .*"))
		qt.Assert(t, qt.IsNil(cfg))
	})

	t.Run("fails to read wrong file type", func(t *testing.T) {
		cfg, err := Parse[TestStruct]("testdata/gopher.png")
		qt.Assert(t, qt.IsNotNil(err))
		qt.Assert(t, qt.ErrorMatches(err, "unmarshalling config: yaml: .*"))
		qt.Assert(t, qt.IsNil(cfg))
	})

	t.Run("fails to validate config struct", func(t *testing.T) {
		cfg, err := Parse[TestStructWithFailingValidation]("testdata/test1.yaml")
		qt.Assert(t, qt.IsNotNil(err))
		qt.Assert(t, qt.ErrorMatches(err, "validating config: this is going to fail"))
		qt.Assert(t, qt.IsNil(cfg))
	})
}

func TestUnmarshalConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		b := []byte(`name: test`)
		cfg := struct {
			Name string `yaml:"name"`
		}{}
		if err := UnmarshalConfig(&cfg, b); err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.Name, "test"))
	})

	t.Run("valid config with $ inside string", func(t *testing.T) {
		t.Setenv("NAME", "testing")

		b := []byte(`name: $NAME`)
		cfg := struct {
			Name string `yaml:"name"`
		}{}
		if err := UnmarshalConfig(&cfg, b); err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.Name, "$NAME"))
	})

	t.Run("valid config with ${ENV_VAR}", func(t *testing.T) {
		t.Setenv("ENV_VAR", "testing")

		b := []byte(`name: ${ENV_VAR}`)
		cfg := struct {
			Name string `yaml:"name"`
		}{}
		if err := UnmarshalConfig(&cfg, b); err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.Name, "testing"))
	})

	t.Run("valid config fallbacks to ENV_VAR default", func(t *testing.T) {
		b := []byte(`name: ${ENV_VAR:default}`)
		cfg := struct {
			Name string `yaml:"name"`
		}{}
		if err := UnmarshalConfig(&cfg, b); err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.Name, "default"))
	})

	t.Run("valid config with env var in middle of string", func(t *testing.T) {
		b := []byte(`name: "This should return ${ENV_VAR:default}"`)
		cfg := struct {
			Name string `yaml:"name"`
		}{}
		if err := UnmarshalConfig(&cfg, b); err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.Name, "This should return default"))
	})

	t.Run("valid config with multiple env vars", func(t *testing.T) {
		b := []byte(`name: "${FIRST:default1} ${SECOND:default2}"`)
		cfg := struct {
			Name string `yaml:"name"`
		}{}
		if err := UnmarshalConfig(&cfg, b); err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.Name, "default1 default2"))
	})

	t.Run("should ignore dollar var in string", func(t *testing.T) {
		b := []byte(`name: "${FIRST:default1} $thisisignored"`)
		cfg := struct {
			Name string `yaml:"name"`
		}{}
		if err := UnmarshalConfig(&cfg, b); err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.Name, "default1 $thisisignored"))
	})

	t.Run("randomly generated string with $", func(t *testing.T) {
		b := []byte(`password: "my$password123"`)
		cfg := struct {
			Password string `yaml:"password"`
		}{}
		if err := UnmarshalConfig(&cfg, b); err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.Password, "my$password123"))
	})

	t.Run("url with $ symbole", func(t *testing.T) {
		b := []byte(`url: "http://example.com/$path"`)
		cfg := struct {
			Url string `yaml:"url"`
		}{}
		if err := UnmarshalConfig(&cfg, b); err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.Url, "http://example.com/$path"))
	})

	t.Run("parses mixed input", func(t *testing.T) {
		b := []byte(`mixed: "prefix${ENV_VAR:default}$suffix"`)
		cfg := struct {
			Mixed string `yaml:"mixed"`
		}{}
		if err := UnmarshalConfig(&cfg, b); err != nil {
			t.Fatal(err)
		}

		qt.Assert(t, qt.Equals(cfg.Mixed, "prefixdefault$suffix"))
	})

	t.Run("invalid config", func(t *testing.T) {
		b := []byte(`asdasdasdad******`)

		cfg := struct {
			Name string `yaml:"name"`
		}{}
		if err := UnmarshalConfig(&cfg, b); err == nil {
			t.Fatal("expected error")
		}
	})
}
