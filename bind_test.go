package flagged

import (
	"flag"
	"io"
	"strings"
	"testing"
	"time"

	errs "github.com/gomatic/go-error"
	"github.com/stretchr/testify/require"
)

// errBadValue is the sentinel a test flag.Value rejects "bad" with.
const errBadValue errs.Const = "bad value"

// stringSet is a struct-typed flag.Value used to exercise the flag.Value path
// (and confirm such a struct is bound, not recursed into).
type stringSet struct {
	values []string
}

func (s stringSet) String() string { return strings.Join(s.values, ",") }

func (s *stringSet) Set(v string) error {
	if v == "bad" {
		return errBadValue
	}
	s.values = append(s.values, v)
	return nil
}

// bindTo binds target to a fresh, quiet FlagSet.
func bindTo(t *testing.T, target any) (*flag.FlagSet, error) {
	t.Helper()
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.SetOutput(io.Discard)
	return set, Bind(set, target)
}

func TestBind_defaultsFromValueTag(t *testing.T) {
	try := require.New(t)
	var s struct {
		Str  string        `usage:"s" value:"hello"`
		Bool bool          `usage:"b" value:"true"`
		F64  float64       `usage:"f" value:"3.5"`
		I    int           `usage:"i" value:"7"`
		I64  int64         `usage:"i64" value:"-9"`
		U    uint          `usage:"u" value:"11"`
		U64  uint64        `usage:"u64" value:"13"`
		Dur  time.Duration `usage:"d" value:"250ms"`
	}
	set, err := bindTo(t, &s)
	try.NoError(err)
	try.NoError(set.Parse(nil))

	try.Equal("hello", s.Str)
	try.True(s.Bool)
	try.InEpsilon(3.5, s.F64, 1e-9)
	try.Equal(7, s.I)
	try.Equal(int64(-9), s.I64)
	try.Equal(uint(11), s.U)
	try.Equal(uint64(13), s.U64)
	try.Equal(250*time.Millisecond, s.Dur)
}

func TestBind_emptyDefaultKeepsCurrent(t *testing.T) {
	try := require.New(t)
	s := struct {
		Str  string        `usage:"s"`
		Bool bool          `usage:"b"`
		F64  float64       `usage:"f"`
		I    int           `usage:"i"`
		I64  int64         `usage:"i64"`
		U    uint          `usage:"u"`
		U64  uint64        `usage:"u64"`
		Dur  time.Duration `usage:"d"`
	}{Str: "keep", Bool: true, F64: 1.5, I: 2, I64: 3, U: 4, U64: 5, Dur: time.Second}

	set, err := bindTo(t, &s)
	try.NoError(err)
	try.NoError(set.Parse(nil))

	try.Equal("keep", s.Str)
	try.True(s.Bool)
	try.InEpsilon(1.5, s.F64, 1e-9)
	try.Equal(2, s.I)
	try.Equal(int64(3), s.I64)
	try.Equal(uint(4), s.U)
	try.Equal(uint64(5), s.U64)
	try.Equal(time.Second, s.Dur)
}

func TestBind_invalidDefault(t *testing.T) {
	tests := []struct {
		target  any
		wantErr error
		name    string
	}{
		{&struct {
			B bool `usage:"b" value:"notbool"`
		}{}, ErrInvalidDefault, "bool"},
		{&struct {
			F float64 `usage:"f" value:"x"`
		}{}, ErrInvalidDefault, "float64"},
		{&struct {
			I int `usage:"i" value:"x"`
		}{}, ErrInvalidDefault, "int"},
		{&struct {
			I int64 `usage:"i" value:"x"`
		}{}, ErrInvalidDefault, "int64"},
		{&struct {
			U uint `usage:"u" value:"x"`
		}{}, ErrInvalidDefault, "uint"},
		{&struct {
			U uint64 `usage:"u" value:"x"`
		}{}, ErrInvalidDefault, "uint64"},
		{&struct {
			D time.Duration `usage:"d" value:"x"`
		}{}, ErrInvalidDefault, "duration"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bindTo(t, tt.target)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestBind_flagValue(t *testing.T) {
	t.Run("valid default", func(t *testing.T) {
		try := require.New(t)
		var s struct {
			Tags stringSet `usage:"tags" value:"a"`
		}
		set, err := bindTo(t, &s)
		try.NoError(err)
		try.NoError(set.Parse([]string{"-tags", "b"}))
		try.Equal([]string{"a", "b"}, s.Tags.values)
	})

	t.Run("empty default registers without Set", func(t *testing.T) {
		try := require.New(t)
		var s struct {
			Tags stringSet `usage:"tags"`
		}
		set, err := bindTo(t, &s)
		try.NoError(err)
		try.NotNil(set.Lookup("tags"))
		try.Nil(s.Tags.values)
	})

	t.Run("bad default", func(t *testing.T) {
		var s struct {
			Tags stringSet `usage:"tags" value:"bad"`
		}
		_, err := bindTo(t, &s)
		require.ErrorIs(t, err, ErrInvalidDefault)
	})
}
