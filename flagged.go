package flagged

import (
	"flag"
	"os"
	"reflect"
	"strings"
	"unicode"
)

// separator joins a nested struct's prefix to its field names.
const separator = "-"

// Domain types for the strings that flow through binding, so no step passes a
// bare primitive.
type (
	// flagName is a resolved command-line flag name.
	flagName string
	// flagUsage is a flag's usage text.
	flagUsage string
	// defaultValue is a flag's default, resolved from the value: and env: tags.
	defaultValue string
	// namePrefix is the dotted-path prefix accumulated from nested structs.
	namePrefix string
	// token is a single unparsed default value handed to a parser.
	token string
)

// Bind registers, on set, a flag for every exported field of the struct pointed
// to by target that carries a usage: tag. See the [package documentation] for
// the tag vocabulary. It returns [ErrNotStructPointer] when target is not a
// non-nil pointer to a struct, and reports the first binding failure otherwise.
//
// [package documentation]: https://pkg.go.dev/github.com/gomatic/go-flagged
func Bind(set *flag.FlagSet, target any) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return ErrNotStructPointer
	}
	return bindStruct(set, v.Elem(), "")
}

// bindStruct registers every field of a struct value, prefixing names with
// prefix, and stops at the first error.
func bindStruct(set *flag.FlagSet, structValue reflect.Value, prefix namePrefix) error {
	structType := structValue.Type()
	for i := range structType.NumField() {
		if err := bindField(set, structValue.Field(i), structType.Field(i), prefix); err != nil {
			return err
		}
	}
	return nil
}

// bindField registers a single field: it recurses into nested structs, skips
// untagged fields, and rejects tagged unexported fields.
func bindField(set *flag.FlagSet, field reflect.Value, meta reflect.StructField, prefix namePrefix) error {
	if !meta.IsExported() {
		if _, tagged := meta.Tag.Lookup("usage"); tagged {
			return ErrUnexportedField
		}
		return nil
	}
	if field.Kind() == reflect.Struct && !isFlagValue(field) {
		return bindStruct(set, field, namePrefix(prefixed(prefix, meta)))
	}
	usage, ok := meta.Tag.Lookup("usage")
	if !ok {
		return nil
	}
	return bindNames(set, field, meta, prefix, flagUsage(usage))
}

// bindNames registers the field under each of its resolved names, sharing one
// resolved default value.
func bindNames(
	set *flag.FlagSet,
	field reflect.Value,
	meta reflect.StructField,
	prefix namePrefix,
	usage flagUsage,
) error {
	value := resolveDefault(meta)
	for _, name := range flagNames(meta, prefix) {
		if err := bind(set, field, name, usage, value); err != nil {
			return err
		}
	}
	return nil
}

// resolveDefault returns the value: tag, overridden by the env: environment
// variable when it is set.
func resolveDefault(meta reflect.StructField) defaultValue {
	value := meta.Tag.Get("value")
	if env := meta.Tag.Get("env"); env != "" {
		if override, ok := os.LookupEnv(env); ok {
			value = override
		}
	}
	return defaultValue(value)
}

// flagNames returns the flag names for a field: the comma-separated entries of
// the flag: tag (with "_" and empty entries derived from the field name), or the
// single derived name when the tag is absent.
func flagNames(meta reflect.StructField, prefix namePrefix) []flagName {
	tag := meta.Tag.Get("flag")
	if tag == "" {
		return []flagName{prefixed(prefix, meta)}
	}
	parts := strings.Split(tag, ",")
	names := make([]flagName, 0, len(parts))
	for _, part := range parts {
		names = append(names, resolveName(flagName(strings.TrimSpace(part)), meta, prefix))
	}
	return names
}

// resolveName returns an explicit flag name as-is, deriving from the field name
// when the entry is empty or "_".
func resolveName(part flagName, meta reflect.StructField, prefix namePrefix) flagName {
	if part == "" || part == "_" {
		return prefixed(prefix, meta)
	}
	return part
}

// prefixed derives a field's kebab-cased flag name, joined to any prefix.
func prefixed(prefix namePrefix, meta reflect.StructField) flagName {
	derived := deCamel(meta)
	if prefix == "" {
		return derived
	}
	return flagName(string(prefix) + separator + string(derived))
}

// isFlagValue reports whether a field's address implements [flag.Value]. Every
// field reached from [Bind]'s struct pointer is addressable, so Addr is safe.
func isFlagValue(field reflect.Value) bool {
	_, ok := field.Addr().Interface().(flag.Value)
	return ok
}

// deCamel converts a CamelCase struct-field name to a kebab-case flag name.
func deCamel(meta reflect.StructField) flagName {
	out := make([]rune, 0, len(meta.Name)*2)
	for i, r := range meta.Name {
		if i > 0 && unicode.IsUpper(r) {
			out = append(out, rune(separator[0]))
		}
		out = append(out, unicode.ToLower(r))
	}
	return flagName(out)
}
