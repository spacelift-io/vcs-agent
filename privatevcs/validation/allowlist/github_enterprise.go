package allowlist

import (
	"net/http"
	"regexp"

	"github.com/spacelift-io/vcs-agent/privatevcs/validation"
)

type githubEnterprisePattern struct {
	Method string
	Path   *regexp.Regexp
}

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
		Path:   validation.GitHubTarballRegex,
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
	"Get Pull Request": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/(api/v3/)?repos/(?P<project>[^/]+/[^/]+)/pulls/[^/]+$"),
	},
	"Refresh Access Token": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/(api/v3/)?app/installations/[^/]+/access_tokens$"),
	},
	"List Installations": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/(api/v3/)?app/installations$"),
	},
	"Get app details": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/(api/v3/)?app$"),
	},
	"Get user details": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/(api/v3/)?users/[^/]+$"),
	},
}

func matchGitHubEnterpriseRequest(r *http.Request) (string, string, error) {
	for name, pattern := range githubEnterprisePatterns {
		if r.Method != pattern.Method {
			continue
		}

		if matches := pattern.Path.FindStringSubmatch(r.URL.EscapedPath()); matches != nil {
			var project string
			if index := pattern.Path.SubexpIndex("project"); index != -1 {
				project = matches[index]
			}

			return name, project, nil
		}
	}
	return "", "", ErrNoMatch
}
