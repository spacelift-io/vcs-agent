package blocklist

import (
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/spacelift-io/spcontext"
	"gopkg.in/yaml.v3"

	"github.com/spacelift-io/vcs-agent/privatevcs/validation"
)

// List is an explicit list of blocked VCS request patterns.
type List struct {
	Version string `json:"version"`
	Rules   []Rule `json:"rules"`
}

// Load loads a blocklist from a YAML file and validates it.
func Load(path string) (*List, error) {
	var out List

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't read the blocklist file %q", path)
	}

	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, errors.Wrapf(err, "couldn't parse the blocklist file %q", path)
	}

	if err := out.Compile(); err != nil {
		return nil, errors.Wrapf(err, "invalid blocklist file %q", path)
	}

	return &out, nil
}

// Compile compiles the blocklist.
func (l List) Compile() error {
	names := make(map[string]struct{})

	for i, rule := range l.Rules {
		if _, ok := names[rule.Name]; ok {
			return errors.Errorf("duplicate rule name %q", rule.Name)
		}
		names[rule.Name] = struct{}{}

		if err := rule.validate(); err != nil {
			return errors.Wrapf(err, "invalid rule %d", i)
		}
	}

	return nil
}

// Validate validates the request against the blocklist. If the request matches
// any rule, it returns an error.
func (l List) Validate(ctx *spcontext.Context, _ validation.Vendor, r *http.Request) (*spcontext.Context, error) {
	for _, rule := range l.Rules {
		if rule.matches(r) {
			return ctx.With("blocked_by", rule.Name), errors.Errorf("request blocked by rule %q", rule.Name)
		}
	}

	return ctx, nil
}
