// Package yamlcfg configures Go applications from YAML files with
// environment variable expansion, validation, and strict decoding.
//
// The entry points are Parse, ParseWithConfig, ParseFS, and
// UnmarshalConfig. Values of the form ${ENV_NAME} or ${ENV_NAME:default}
// in YAML values are replaced from the process environment before
// decoding. Bare $VAR references are left untouched.
//
// Decoding is strict by default: unknown fields and duplicate mapping keys
// are rejected. This can be relaxed or extended per call with Options such
// as AllowUnknownFields, AllowDuplicateKeys, AllowFieldPrefixes, and
// UseOrderedMap.
//
// Anchors and aliases may be defined in other YAML files by passing the
// WithReferenceFiles, WithReferenceDirs, WithRecursiveDirs,
// WithReferenceReaders, or WithEmbeddedReferences options.
//
// If a config type implements Validate() error, it is called automatically
// after unmarshalling. Alternatively, or additionally, a StructValidator
// (as implemented by validation libraries such as go-playground/validator)
// can be plugged in with WithValidator.
//
// YAML errors returned by this package wrap the typed errors of the
// underlying github.com/goccy/go-yaml decoder. They can be inspected with
// errors.As, the IsSyntaxError, IsTypeError, IsUnknownFieldError,
// IsDuplicateKeyError, and IsOverflowError predicates, and the Position
// helper.
package yamlcfg
