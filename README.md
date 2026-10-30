# yamlcfg

yamlcfg is a wrapper around the [goccy/go-yaml](https://github.com/goccy/go-yaml) library and provides a convenient way to configure Go applications with YAML and environment variables.

The library can also automatically call `Validate` functions if present on the given config struct.

## Why?

Not every configuration value needs to be an environment variable. Database pool sizes, timeouts, feature flags, and other stable settings are better expressed as static YAML values that live in version control alongside your code. But secrets and environment-specific values (database URLs, API keys, listen addresses) still need to come from the environment.

yamlcfg lets you combine both in a single config file — static values that rarely change sit next to `${VAR}` references that get resolved at startup. This means fewer environment variables to manage, sensible defaults via `${VAR:default}` syntax, and a single file that documents your application's full configuration surface.

## Installation

```
go get github.com/aranw/yamlcfg
```

## Environment Variables

yamlcfg expands environment variables in YAML values using the `${VAR}` syntax:

```yaml
database_url: "${DATABASE_URL}"
```

Default values are supported with `${VAR:default}`:

```yaml
log_level: "${LOG_LEVEL:info}"
port: "${PORT:8080}"
```

Environment variables can also appear within larger strings:

```yaml
dsn: "host=${DB_HOST:localhost} port=${DB_PORT:5432} dbname=${DB_NAME:myapp}"
```

Note: bare `$VAR` references (without braces) are treated as literal strings and not expanded.

## Usage

### Parse

Reads a YAML file from disk and unmarshals it into the given type:

```go
type Config struct {
    LogLevel string `yaml:"log_level"`
    Port     int    `yaml:"port"`
}

cfg, err := yamlcfg.Parse[Config]("config.yaml")
if err != nil {
    log.Fatal(err)
}
```

### ParseWithConfig

Parses a YAML file with pre-populated default values. Fields not present in the YAML file retain their defaults:

```go
defaults := &Config{
    LogLevel: "info",
    Port:     8080,
}

cfg, err := yamlcfg.ParseWithConfig(defaults, "config.yaml")
if err != nil {
    log.Fatal(err)
}
```

### ParseFS

Reads a YAML file from an `embed.FS`:

```go
//go:embed config.yaml
var configFS embed.FS

cfg, err := yamlcfg.ParseFS[Config](configFS, "config.yaml")
if err != nil {
    log.Fatal(err)
}
```

### UnmarshalConfig

Unmarshals raw YAML bytes directly, expanding environment variables:

```go
data := []byte(`log_level: "${LOG_LEVEL:debug}"`)

var cfg Config
if err := yamlcfg.UnmarshalConfig(&cfg, data); err != nil {
    log.Fatal(err)
}
```

## Validation

If your config struct implements a `Validate() error` method, it will be called automatically after unmarshalling:

```go
type Config struct {
    LogLevel string `yaml:"log_level"`
}

func (c *Config) Validate() error {
    validLevels := []string{"debug", "info", "warn", "error"}
    if !slices.Contains(validLevels, c.LogLevel) {
        return fmt.Errorf("invalid log level: %s", c.LogLevel)
    }
    return nil
}
```

## Options

All loading functions accept functional `Option`s. Decoding is strict by default (unknown fields and duplicate keys are rejected); options can relax or extend that behaviour:

```go
cfg, err := yamlcfg.Parse[Config]("config.yaml",
    yamlcfg.AllowUnknownFields(),          // ignore unknown keys
    yamlcfg.AllowDuplicateKeys(),          // last duplicate key wins
    yamlcfg.AllowFieldPrefixes("x_"),      // let strict mode pass keys with these prefixes
    yamlcfg.UseOrderedMap(),               // decode untyped maps as *yaml.MapSlice
)
```

## Cross-file references

Anchors and aliases can be defined in other YAML files, enabling shared `defaults.yaml`-style configs:

```go
cfg, err := yamlcfg.Parse[Config]("config.yaml",
    yamlcfg.WithReferenceFiles("defaults.yaml"),
)
```

`WithReferenceDirs` and `WithRecursiveDirs` load reference files from directories, `WithReferenceReaders` from any `io.Reader`, and `WithEmbeddedReferences` from the same `embed.FS` used with `ParseFS`:

```go
cfg, err := yamlcfg.ParseFS[Config](configFS, "config.yaml",
    yamlcfg.WithEmbeddedReferences(configFS, "defaults.yaml"),
)
```

Note: `${VAR}` environment variable expansion only applies to the main config data, not to referenced files.

## Pluggable validation

Besides the `Validate() error` interface, a `StructValidator` can be plugged in — the interface implemented by [go-playground/validator](https://github.com/go-playground/validator), so tag-based validation runs automatically after decoding:

```go
validate := validator.New()

cfg, err := yamlcfg.Parse[Config]("config.yaml",
    yamlcfg.WithValidator(validate), // struct tags like `validate:"min=1"`
)
```

## Typed errors

YAML errors returned by this package wrap the underlying typed errors from goccy/go-yaml, so failures can be handled programmatically with `errors.As`:

```go
var fieldErr *yamlcfg.UnknownFieldError
if errors.As(err, &fieldErr) {
    line, col, _ := yamlcfg.Position(err)
    log.Printf("unknown field at config.yaml:%d:%d", line, col)
}
```

Convenience predicates `IsSyntaxError`, `IsTypeError`, `IsUnknownFieldError`, `IsDuplicateKeyError` and `IsOverflowError` are also available, and every typed error's `Error()` includes the offending source line with a position marker.

## License

The yamlcfg package is licensed under the MIT license. Please see the LICENSE file for details.
