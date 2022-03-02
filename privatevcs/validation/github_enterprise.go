package validation

import (
	"net/http"
	"regexp"
	"strings"
)

type githubEnterprisePattern struct {
	Method string
	Path   *regexp.Regexp
}

var tarballRegex *regexp.Regexp = regexp.MustCompile("^/(_?codeload/)?(?P<project>[^/]+/[^/]+)/legacy.tar.gz/[^/]+$")

var githubEnterprisePatterns = map[string]githubEnterprisePattern{
	"Compare Trees": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/(api/v3/)?repos/(?P<project>[^/]+/[^/]+)/compare/[^/]...[^/]+"),
	},
	"Create Commit Status": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/(api/v3/)?repos/(?P<project>[^/]+/[^/]+)/statuses/[^/]+$"),
	},
	"Create Check Run": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/(api/v3/)?repos/(?P<project>[^/]+/[^/]+)/check-runs$"),
	},
	"Create Deployment": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/(api/v3/)?repos/(?P<project>[^/]+/[^/]+)/deployments$"),
	},
	"Create Deployment Status": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/(api/v3/)?repos/(?P<project>[^/]+/[^/]+)/deployments/[^/]+/statuses$"),
	},
	"Delete Deployment": {
		Method: http.MethodDelete,
		Path:   regexp.MustCompile("^/(api/v3/)?repos/(?P<project>[^/]+/[^/]+)/deployments/[^/]+$"),
	},
	"Get Individual Commit Details": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/(api/v3/)?repos/(?P<project>[^/]+/[^/]+)/commits/[^/]+"),
	},
	"Get Repository Tarball": {
		Method: http.MethodGet,
		Path:   tarballRegex,
	},
	"GraphQL Endpoint": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/(api/)?graphql$"),
	},
	"List Deployments": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/(api/v3/)?repos/(?P<project>[^/]+/[^/]+)/deployments$"),
	},
	"List Pull Request Files": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/(api/v3/)?repos/(?P<project>[^/]+/[^/]+)/pulls/[^/]+/files$"),
	},
	"Refresh Access Token": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/(api/v3/)?app/installations/[^/]+/access_tokens$"),
	},
	"List Installations": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/(api/v3/)?app/installations$"),
	},
}

func matchGitHubEnterpriseRequest(r *http.Request) (string, string, *string, error) {
	for name, pattern := range githubEnterprisePatterns {
		if matches := pattern.Path.FindStringSubmatch(r.URL.EscapedPath()); matches != nil {
			var project string
			if index := pattern.Path.SubexpIndex("project"); index != -1 {
				project = matches[index]
			}

			_, subdomain := IsGitHubTarballRequest(r)

			return name, project, subdomain, nil
		}
	}
	return "", "", nil, ErrNoMatch
}

// IsGitHubTarballRequest returns whether the request is a GitHub Enterprise tarball download request.
// If it is a download request, and if the server is using subdomain isolation, subdomain
// will contain the subdomain to prefix the request hostname with.
func IsGitHubTarballRequest(r *http.Request) (ok bool, subdomain *string) {
	// If we're trying to download the source, and the path doesn't start with "/_codeload"
	// or "/codeload" the GHE instance must have subdomain isolation enabled, so we need to prefix the
	// hostname with "codeload."
	if tarballRegex.MatchString(r.URL.EscapedPath()) {
		if !strings.HasPrefix(r.URL.EscapedPath(), "/_codeload") &&
			!strings.HasPrefix(r.URL.EscapedPath(), "/codeload") {
			h := "codeload"
			subdomain = &h
		}

		return true, subdomain
	}

	return false, nil
}
