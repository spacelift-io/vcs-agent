package blocklist

import (
	"net/http"
	"regexp"

	"github.com/pkg/errors"
	"github.com/spacelift-io/vcs-agent/privatevcs/validation"
)

// Rule is a single rule in a blocklist.
type Rule struct {
	// The name of the rule.
	Name string `json:"name"`

	// The HTTP method to match.
	Method string `json:"method"`

	// The regular expression to match the path against.
	Path string `json:"path"`

	// The regular expression to match the vendor against.
	Vendor string `json:"vendor"`

	methodRegexp, pathRegexp, vendorRegexp *regexp.Regexp
}

func (r *Rule) compile() error {
	var err error
	r.methodRegexp, err = regexp.Compile(r.Method)
	if err != nil {
		return errors.Wrapf(err, "couldn't compile method regexp for rule %q", r.Name)
	}

	r.pathRegexp, err = regexp.Compile(r.Path)
	if err != nil {
		return errors.Wrapf(err, "couldn't compile path regexp for rule %q", r.Name)
	}

	r.vendorRegexp, err = regexp.Compile(r.Vendor)
	return errors.Wrapf(err, "couldn't compile vendor regexp for rule %q", r.Name)
}

func (r *Rule) matches(vendor validation.Vendor, req *http.Request) bool {
	if !r.vendorRegexp.MatchString(string(vendor)) {
		return false
	}

	if !r.methodRegexp.MatchString(req.Method) {
		return false
	}

	return r.pathRegexp.MatchString(req.URL.Path)
}

func (r *Rule) validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}

	return errors.Wrapf(r.compile(), "couldn't validate rule %q", r.Name)
}
