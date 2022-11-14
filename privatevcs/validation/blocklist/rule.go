package blocklist

import (
	"net/http"
	"regexp"

	"github.com/pkg/errors"
)

// Rule is a single rule in a blocklist.
type Rule struct {
	// The name of the rule.
	Name string `json:"name"`

	// The HTTP method to match.
	Method string `json:"method"`

	// The regular expression to match the path against.
	Path string `json:"path"`

	methodRegexp, pathRegexp *regexp.Regexp
}

// Matches returns true if the rule matches the given request.
func (r *Rule) Matches(req *http.Request) bool {
	return r.methodRegexp.MatchString(req.Method) && r.pathRegexp.MatchString(req.URL.Path)
}

// Validate compiles and validates the rule.
func (r *Rule) Validate() error {
	if r.Name == "" {
		return errors.New("rule name is required")
	}

	return errors.Wrapf(r.compile(), "could not compile rule %q", r.Name)
}

func (r *Rule) compile() error {
	var err error
	r.methodRegexp, err = regexp.Compile(r.Method)
	if err != nil {
		return errors.Wrapf(err, "invalid method matcher")
	}

	r.pathRegexp, err = regexp.Compile(r.Path)
	return errors.Wrapf(err, "invalid path matcher")
}
