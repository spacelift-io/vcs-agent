package validation

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/spacelift-io/spcontext"
	"github.com/spacelift-io/vcs-agent/nullable"
)

// GitHubTarballRegex is a regular expression that matches GitHub Enterprise
// tarball download requests.
var GitHubTarballRegex *regexp.Regexp = regexp.MustCompile("^/(_?codeload/)?(?P<project>[^/]+/[^/]+)/legacy.tar.gz/[^/]+$")

// IsGitHubTarballRequest returns whether the request is a GitHub Enterprise tarball download request.
// If it is a download request, and if the server is using subdomain isolation, subdomain
// will contain the subdomain to prefix the request hostname with.
func IsGitHubTarballRequest(r *http.Request) (ok bool, subdomain *string) {
	// If we're trying to download the source, and the path doesn't start with "/_codeload"
	// or "/codeload" the GHE instance must have subdomain isolation enabled, so we need to prefix the
	// hostname with "codeload."
	if GitHubTarballRegex.MatchString(r.URL.EscapedPath()) {
		if !strings.HasPrefix(r.URL.EscapedPath(), "/_codeload") &&
			!strings.HasPrefix(r.URL.EscapedPath(), "/codeload") {
			subdomain = nullable.String("codeload")
		}

		return true, subdomain
	}

	return false, nil
}

// RewriteGitHubTarballRequest rewrites the GitHub tarball request to use
// the right subdomain, if necessary.
func RewriteGitHubTarballRequest(ctx *spcontext.Context, vendor Vendor, req *http.Request) *spcontext.Context {
	if vendor != GitHubEnterprise {
		return ctx
	}

	_, subdomain := IsGitHubTarballRequest(req)
	if subdomain == nil {
		return ctx
	}

	req.URL.Host = *subdomain + "." + req.URL.Host
	req.Host = req.URL.Host

	return ctx.With("subdomain", *subdomain)
}
