package flagged

import (
	"flag"
	"reflect"
	"strconv"
	"time"
)

// bind registers field on set under name, using raw as the default (empty means
// "keep the field's current value"). It dispatches on the field's concrete
// pointer type — kept local so no flag-target pointer is passed as a parameter —
// and rejects a name already present on the set.
func bind(set *flag.FlagSet, field reflect.Value, name flagName, usage flagUsage, raw defaultValue) error {
	if set.Lookup(string(name)) != nil {
		return ErrDuplicateFlag.With(nil, string(name))
	}
	switch p := field.Addr().Interface().(type) {
	case flag.Value:
		return bindValue(set, p, name, usage, raw)
	case *string:
		return register(raw, *p, parseString, func(v string) { set.StringVar(p, string(name), v, string(usage)) })
	case *bool:
		return register(raw, *p, parseBool, func(isEnabled bool) { set.BoolVar(p, string(name), isEnabled, string(usage)) })
	case *int:
		return register(raw, *p, parseInt, func(v int) { set.IntVar(p, string(name), v, string(usage)) })
	case *int64:
		return register(raw, *p, parseInt64, func(v int64) { set.Int64Var(p, string(name), v, string(usage)) })
	case *uint:
		return register(raw, *p, parseUint, func(v uint) { set.UintVar(p, string(name), v, string(usage)) })
	case *uint64:
		return register(raw, *p, parseUint64, func(v uint64) { set.Uint64Var(p, string(name), v, string(usage)) })
	case *float64:
		return register(raw, *p, parseFloat64, func(v float64) { set.Float64Var(p, string(name), v, string(usage)) })
	case *time.Duration:
		return register(raw, *p, parseDuration, func(v time.Duration) { set.DurationVar(p, string(name), v, string(usage)) })
	default:
		return ErrUnsupportedType.With(nil, field.Type().String())
	}
}

// register resolves the default value for a flag and hands it to assign: it uses
// current when raw is empty, otherwise parses raw and reports a parse failure as
// [ErrInvalidDefault].
func register[T any](raw defaultValue, current T, parse func(token) (T, error), assign func(T)) error {
	if raw == "" {
		assign(current)
		return nil
	}
	parsed, err := parse(token(raw))
	if err != nil {
		return ErrInvalidDefault.With(err, string(raw))
	}
	assign(parsed)
	return nil
}

// bindValue registers a field whose address implements [flag.Value], seeding its
// default from raw when present.
func bindValue(set *flag.FlagSet, value flag.Value, name flagName, usage flagUsage, raw defaultValue) error {
	if raw != "" {
		if err := value.Set(string(raw)); err != nil {
			return ErrInvalidDefault.With(err, string(raw))
		}
	}
	set.Var(value, string(name), string(usage))
	return nil
}

func parseString(t token) (string, error) { return string(t), nil }

func parseBool(t token) (bool, error) { return strconv.ParseBool(string(t)) }

func parseInt(t token) (int, error) { return strconv.Atoi(string(t)) }

func parseInt64(t token) (int64, error) { return strconv.ParseInt(string(t), 10, 64) }

func parseUint(t token) (uint, error) {
	v, err := strconv.ParseUint(string(t), 10, strconv.IntSize)
	return uint(v), err
}

func parseUint64(t token) (uint64, error) { return strconv.ParseUint(string(t), 10, 64) }

func parseFloat64(t token) (float64, error) { return strconv.ParseFloat(string(t), 64) }

func parseDuration(t token) (time.Duration, error) { return time.ParseDuration(string(t)) }
