package allowlist

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/pkg/errors"
	"github.com/spacelift-io/spcontext"

	"github.com/spacelift-io/vcs-agent/privatevcs/validation"
)

// List implements a validation strategy based on an explicit, exhaustive,
// code-based allowlist (the legacy approach).
type List struct {
	projectRegexp *regexp.Regexp
}

// New creates a new allowlist strategy from a project regexp.
func New(projectRegexp string) (*List, error) {
	r, err := regexp.Compile(projectRegexp)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't compile project regexp %q", projectRegexp)
	}

	return &List{projectRegexp: r}, nil
}

// Validate validates the request and returns an error if the request should be
// blocked.
func (l List) Validate(ctx *spcontext.Context, vendor validation.Vendor, req *http.Request) (*spcontext.Context, error) {
	name, project, err := MatchRequest(vendor, req)
	if err != nil {
		return ctx.With("match_error", err), ctx.RawError(err, "invalid request")
	}

	ctx = ctx.With("name", name)

	projectUnescaped, err := url.PathUnescape(project)
	if err != nil {
		ctx := ctx.With(
			"match_error", err,
			"project_urlencoded", project,
		)
		return ctx, ctx.RawError(err, "couldn't url-unescape project name")
	}

	if project != "" && !l.projectRegexp.MatchString(projectUnescaped) {
		ctx := ctx.With(
			"match_error", err,
			"project", projectUnescaped,
			"project_regexp", l.projectRegexp.String(),
		)

		return ctx, ctx.RawError(fmt.Errorf("request project didn't match allowed projects regexp"), "invalid request")
	}

	return ctx.With("project", project), nil
}
