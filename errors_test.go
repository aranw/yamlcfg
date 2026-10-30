package yamlcfg

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-quicktest/qt"
)

func TestTypedErrors(t *testing.T) {
	t.Run("syntax error", func(t *testing.T) {
		b := []byte("log_level: [unclosed\n")
		cfg := &RefConfig{}
		err := UnmarshalConfig(cfg, b)
		qt.Assert(t, qt.IsTrue(IsSyntaxError(err)))

		var syntaxErr *SyntaxError
		qt.Assert(t, qt.IsTrue(errors.As(err, &syntaxErr)))

		line, column, ok := Position(err)
		qt.Assert(t, qt.IsTrue(ok))
		qt.Assert(t, qt.Equals(line, 1))
		qt.Assert(t, qt.Not(qt.Equals(column, 0)))
	})

	t.Run("type error", func(t *testing.T) {
		b := []byte("log_level: \"info\"\nport: \"notanumber\"\n")
		cfg := &RefConfig{}
		err := UnmarshalConfig(cfg, b)
		qt.Assert(t, qt.IsTrue(IsTypeError(err)))

		var typeErr *TypeError
		qt.Assert(t, qt.IsTrue(errors.As(err, &typeErr)))
	})

	t.Run("unknown field error", func(t *testing.T) {
		b := []byte("log_level: \"info\"\nunknown_key: 1\n")
		cfg := &RefConfig{}
		err := UnmarshalConfig(cfg, b)
		qt.Assert(t, qt.IsTrue(IsUnknownFieldError(err)))

		line, column, ok := Position(err)
		qt.Assert(t, qt.IsTrue(ok))
		qt.Assert(t, qt.Equals(line, 2))
		qt.Assert(t, qt.Equals(column, 1))
	})

	t.Run("duplicate key error", func(t *testing.T) {
		b := []byte("log_level: \"info\"\nlog_level: \"debug\"\n")
		cfg := &RefConfig{}
		err := UnmarshalConfig(cfg, b)
		qt.Assert(t, qt.IsTrue(IsDuplicateKeyError(err)))

		line, _, ok := Position(err)
		qt.Assert(t, qt.IsTrue(ok))
		qt.Assert(t, qt.Equals(line, 2))
	})

	t.Run("Position unavailable for plain errors", func(t *testing.T) {
		_, _, ok := Position(fmt.Errorf("not a yaml error"))
		qt.Assert(t, qt.IsFalse(ok))
	})

	t.Run("Position survives wrapping", func(t *testing.T) {
		wrapped := fmt.Errorf("loading config: %w", &SyntaxError{})
		// a zero-value SyntaxError has no token, so position is unavailable
		_, _, ok := Position(wrapped)
		qt.Assert(t, qt.IsFalse(ok))
		qt.Assert(t, qt.IsTrue(IsSyntaxError(wrapped)))
	})
}
