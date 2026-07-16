// Package flagged registers command-line flags declaratively from struct tags.
//
// [Bind] walks a pointer to a struct and registers a [flag.FlagSet] flag for
// each exported field that carries a usage: tag. The flag's name comes from the
// flag: tag (a comma-separated primary name and aliases), or the kebab-cased
// field name when that tag is absent; its default comes from the value: tag,
// which a set env: environment variable overrides. Nested struct fields are
// registered recursively with their names prefixed. The caller owns the
// FlagSet and calls Parse, so flagged holds no global state and never parses on
// its own.
//
// Given a tagged struct:
//
//	var settings struct {
//		Host    string        `usage:"server host"   value:"localhost"`
//		Port    int           `flag:"port,p" usage:"listen port" value:"8080" env:"PORT"`
//		Verbose bool          `flag:"verbose,v" usage:"verbose output"`
//		Timeout time.Duration `usage:"request timeout" value:"5s"`
//	}
//
// register its flags on a set:
//
//	set := flag.NewFlagSet("app", flag.ContinueOnError)
//	if err := flagged.Bind(set, &settings); err != nil {
//		return err
//	}
//	err := set.Parse(os.Args[1:])
package flagged
