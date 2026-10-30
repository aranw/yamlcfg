package yamlcfg

import (
	"errors"
	"strings"

	"github.com/goccy/go-yaml"
)

// Typed YAML errors reported by the underlying goccy/go-yaml decoder. All
// errors returned by this package wrap these where applicable, so they can
// be matched with errors.As:
//
//	var syntaxErr *yamlcfg.SyntaxError
//	if errors.As(err, &syntaxErr) { ... }
type (
	// Error is implemented by every typed error below. It provides the
	// offending token, a plain message, and pretty source-formatted output.
	Error = yaml.Error
	// SyntaxError reports malformed YAML, including duplicate mapping keys.
	SyntaxError = yaml.SyntaxError
	// TypeError reports a YAML value that cannot be converted to the
	// destination Go type.
	TypeError = yaml.TypeError
	// UnknownFieldError reports a YAML key with no matching struct field.
	UnknownFieldError = yaml.UnknownFieldError
	// DuplicateKeyError reports a duplicate mapping key.
	DuplicateKeyError = yaml.DuplicateKeyError
	// OverflowError reports a number that overflows the destination type.
	OverflowError = yaml.OverflowError
)

// IsSyntaxError reports whether err was caused by malformed YAML.
func IsSyntaxError(err error) bool {
	return isYAMLType[*SyntaxError](err)
}

// IsTypeError reports whether err was caused by a YAML value that cannot be
// converted to the destination Go type.
func IsTypeError(err error) bool {
	return isYAMLType[*TypeError](err)
}

// IsUnknownFieldError reports whether err was caused by a YAML key that does
// not match any struct field (only possible without AllowUnknownFields).
func IsUnknownFieldError(err error) bool {
	return isYAMLType[*UnknownFieldError](err)
}

// IsOverflowError reports whether err was caused by a number that overflows
// the destination Go type.
func IsOverflowError(err error) bool {
	return isYAMLType[*OverflowError](err)
}

// IsDuplicateKeyError reports whether err was caused by a duplicate mapping
// key (only possible without AllowDuplicateKeys). goccy currently reports
// duplicate keys as a SyntaxError whose message says the key is "already
// defined", so this check also matches that form.
func IsDuplicateKeyError(err error) bool {
	if isYAMLType[*DuplicateKeyError](err) {
		return true
	}
	var syntaxErr *SyntaxError
	if !errors.As(err, &syntaxErr) {
		return false
	}
	return strings.Contains(syntaxErr.GetMessage(), "already defined")
}

// Position returns the line and column (1-based) of the YAML token that
// caused err, if that information is available.
func Position(err error) (line, column int, ok bool) {
	var yamlErr Error
	if !errors.As(err, &yamlErr) {
		return 0, 0, false
	}
	tk := yamlErr.GetToken()
	if tk == nil {
		return 0, 0, false
	}
	return tk.Position.Line, tk.Position.Column, true
}

func isYAMLType[E Error](err error) bool {
	var e E
	return errors.As(err, &e)
}
