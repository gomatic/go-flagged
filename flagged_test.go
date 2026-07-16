package flagged

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBind_notStructPointer(t *testing.T) {
	tests := []struct {
		target  any
		wantErr error
		name    string
	}{
		{struct {
			X int `usage:"x"`
		}{}, ErrNotStructPointer, "struct value"},
		{(*struct{})(nil), ErrNotStructPointer, "nil pointer"},
		{new(int), ErrNotStructPointer, "pointer to non-struct"},
		{nil, ErrNotStructPointer, "nil interface"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bindTo(t, tt.target)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestBind_names(t *testing.T) {
	try := require.New(t)
	var s struct {
		AString string `usage:"a"`                  // derived: a-string
		Verbose bool   `flag:"verbose,v" usage:"v"` // explicit + alias
		Named   int    `flag:"_,n" usage:"x"`       // derived + alias
		Empty   int    `flag:",e" usage:"x"`        // empty entry -> derived
	}
	set, err := bindTo(t, &s)
	try.NoError(err)

	for _, name := range []string{"a-string", "verbose", "v", "named", "n", "empty", "e"} {
		try.NotNil(set.Lookup(name), name)
	}

	// An alias writes the same field as its primary name.
	try.NoError(set.Parse([]string{"-v"}))
	try.True(s.Verbose)
}

func TestBind_nested(t *testing.T) {
	try := require.New(t)
	var s struct {
		Server struct {
			Host string `usage:"h" value:"localhost"`
			Port int    `usage:"p" value:"80"`
		}
	}
	set, err := bindTo(t, &s)
	try.NoError(err)
	try.NoError(set.Parse(nil))

	try.NotNil(set.Lookup("server-host"))
	try.NotNil(set.Lookup("server-port"))
	try.Equal("localhost", s.Server.Host)
	try.Equal(80, s.Server.Port)
}

func TestBind_envOverride(t *testing.T) {
	try := require.New(t)
	t.Setenv("FLAGGED_TEST_PORT", "9999")
	var s struct {
		Host string `usage:"h" value:"local" env:"FLAGGED_TEST_UNSET"`
		Port int    `usage:"p" value:"8080" env:"FLAGGED_TEST_PORT"`
	}
	set, err := bindTo(t, &s)
	try.NoError(err)
	try.NoError(set.Parse(nil))

	try.Equal(9999, s.Port)    // environment overrides the value tag
	try.Equal("local", s.Host) // unset environment falls back to the value tag
}

func TestBind_untaggedFieldSkipped(t *testing.T) {
	try := require.New(t)
	var s struct {
		Tracked   int `usage:"t"`
		Untracked int
	}
	set, err := bindTo(t, &s)
	try.NoError(err)

	try.NotNil(set.Lookup("tracked"))
	try.Nil(set.Lookup("untracked"))
}

func TestBind_unsupportedType(t *testing.T) {
	_, err := bindTo(t, &struct {
		Names []string `usage:"n"`
	}{})
	require.ErrorIs(t, err, ErrUnsupportedType)
}

func TestBind_duplicateFlag(t *testing.T) {
	_, err := bindTo(t, &struct {
		A int `flag:"dup" usage:"a"`
		B int `flag:"dup" usage:"b"`
	}{})
	require.ErrorIs(t, err, ErrDuplicateFlag)
}

func TestBind_nestedError(t *testing.T) {
	_, err := bindTo(t, &struct {
		Sub struct {
			Bad int `usage:"x" value:"notint"`
		}
	}{})
	require.ErrorIs(t, err, ErrInvalidDefault)
}

// taggedHidden has an unexported field carrying a usage tag, which must be
// rejected rather than bound.
type taggedHidden struct {
	hidden int `usage:"x"`
}

func TestBind_unexportedTaggedField(t *testing.T) {
	target := &taggedHidden{hidden: 1}
	_ = target.hidden // read so the field is not reported unused
	_, err := bindTo(t, target)
	require.ErrorIs(t, err, ErrUnexportedField)
}

// mixedVisibility pairs an unexported untagged field with an exported tagged one.
type mixedVisibility struct {
	Name   string `usage:"n" value:"x"`
	secret int
}

func TestBind_unexportedUntaggedSkipped(t *testing.T) {
	try := require.New(t)
	target := &mixedVisibility{secret: 7}
	_ = target.secret // read so the field is not reported unused
	set, err := bindTo(t, target)
	try.NoError(err)
	try.NotNil(set.Lookup("name"))
}
