package yamlcfg

import (
	"embed"
	"io"
	"io/fs"

	"github.com/goccy/go-yaml"
)

// Option customises how a YAML config is decoded. Options are accepted by
// Parse, ParseWithConfig, ParseFS and UnmarshalConfig.
type Option func(*decodeConfig)

// decodeConfig collects the decoder behaviour selected by Options. Strict
// decoding (unknown fields are rejected) is the default.
type decodeConfig struct {
	strict        bool
	decodeOptions []yaml.DecodeOption
}

func newDecodeConfig(opts []Option) *decodeConfig {
	c := &decodeConfig{strict: true}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *decodeConfig) decoder(r io.Reader) *yaml.Decoder {
	options := make([]yaml.DecodeOption, 0, len(c.decodeOptions)+1)
	if c.strict {
		options = append(options, yaml.Strict())
	}
	options = append(options, c.decodeOptions...)
	return yaml.NewDecoder(r, options...)
}

// StructValidator validates a decoded value. It matches the interface
// implemented by validation libraries such as go-playground/validator
// (see validator.Validate.Struct), which can be plugged in directly.
type StructValidator = yaml.StructValidator

// AllowUnknownFields disables strict decoding so that YAML keys that do not
// match any struct field are ignored instead of rejected.
func AllowUnknownFields() Option {
	return func(c *decodeConfig) { c.strict = false }
}

// AllowDuplicateKeys allows duplicate mapping keys in the YAML input. Without
// this option duplicate keys are an error. When allowed, the last value wins.
func AllowDuplicateKeys() Option {
	return func(c *decodeConfig) {
		c.decodeOptions = append(c.decodeOptions, yaml.AllowDuplicateMapKey())
	}
}

// UseOrderedMap preserves mapping key order by decoding untyped values
// (interface{} destinations) as a *yaml.MapSlice instead of a map.
func UseOrderedMap() Option {
	return func(c *decodeConfig) {
		c.decodeOptions = append(c.decodeOptions, yaml.UseOrderedMap())
	}
}

// AllowFieldPrefixes, while strict decoding is enabled, allows YAML keys with
// the given prefixes to bypass the unknown field check.
func AllowFieldPrefixes(prefixes ...string) Option {
	return func(c *decodeConfig) {
		c.decodeOptions = append(c.decodeOptions, yaml.AllowFieldPrefixes(prefixes...))
	}
}

// WithValidator registers a StructValidator that runs against the decoded
// value before decoding is considered successful.
func WithValidator(v StructValidator) Option {
	return func(c *decodeConfig) {
		c.decodeOptions = append(c.decodeOptions, yaml.Validator(v))
	}
}

// WithReferenceFiles allows the config to reference anchors and aliases
// defined in the given YAML files. Note that environment variable expansion
// only applies to the main config data, not to referenced files.
func WithReferenceFiles(paths ...string) Option {
	return func(c *decodeConfig) {
		c.decodeOptions = append(c.decodeOptions, yaml.ReferenceFiles(paths...))
	}
}

// WithReferenceDirs allows the config to reference anchors and aliases
// defined in YAML files under the given directories (searched
// non-recursively).
func WithReferenceDirs(dirs ...string) Option {
	return func(c *decodeConfig) {
		c.decodeOptions = append(c.decodeOptions, yaml.ReferenceDirs(dirs...))
	}
}

// WithRecursiveDirs works like WithReferenceDirs but searches the given
// directories recursively.
func WithRecursiveDirs(dirs ...string) Option {
	return func(c *decodeConfig) {
		c.decodeOptions = append(c.decodeOptions,
			yaml.ReferenceDirs(dirs...),
			yaml.RecursiveDir(true),
		)
	}
}

// WithReferenceReaders allows the config to reference anchors and aliases
// defined in YAML documents read from the given readers.
func WithReferenceReaders(readers ...io.Reader) Option {
	return func(c *decodeConfig) {
		c.decodeOptions = append(c.decodeOptions, yaml.ReferenceReaders(readers...))
	}
}

// WithEmbeddedReferences allows a config parsed from an embed.FS to reference
// anchors and aliases defined in other YAML files from the same embed.FS.
// Files are opened lazily, so a missing file surfaces as a decode error.
func WithEmbeddedReferences(fsys embed.FS, paths ...string) Option {
	return func(c *decodeConfig) {
		readers := make([]io.Reader, 0, len(paths))
		for _, p := range paths {
			readers = append(readers, &embeddedReader{fsys: fsys, path: p})
		}
		c.decodeOptions = append(c.decodeOptions, yaml.ReferenceReaders(readers...))
	}
}

// embeddedReader lazily opens an embedded file on first Read.
type embeddedReader struct {
	fsys embed.FS
	path string
	file fs.File
}

var _ io.Reader = (*embeddedReader)(nil)

func (r *embeddedReader) Read(p []byte) (int, error) {
	if r.file == nil {
		f, err := r.fsys.Open(r.path)
		if err != nil {
			return 0, err
		}
		r.file = f
	}
	return r.file.Read(p)
}
