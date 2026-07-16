package flagged

import (
	errs "github.com/gomatic/go-error"
)

// Every error the package can emit is a const of the ecosystem's [errs.Const]
// sentinel type, so each is matchable with errors.Is rather than by string.
const (
	// ErrNotStructPointer is returned by [Bind] when the target is not a non-nil
	// pointer to a struct.
	ErrNotStructPointer errs.Const = "target must be a non-nil pointer to a struct"
	// ErrUnexportedField is returned by [Bind] when a field carrying a usage tag
	// cannot be set because it is unexported.
	ErrUnexportedField errs.Const = "cannot bind an unexported field"
	// ErrUnsupportedType is returned by [Bind] when a tagged field's type has no
	// flag binding.
	ErrUnsupportedType errs.Const = "unsupported field type"
	// ErrInvalidDefault is returned by [Bind] when a value or env default cannot
	// be parsed as the field's type.
	ErrInvalidDefault errs.Const = "invalid default value"
	// ErrDuplicateFlag is returned by [Bind] when a flag name is already
	// registered on the set.
	ErrDuplicateFlag errs.Const = "flag already registered"
)
