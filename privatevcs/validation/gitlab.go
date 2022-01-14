package validation

import (
	"net/http"
	"regexp"
)

type gitlabPattern struct {
	Method string
	Path   *regexp.Regexp
}

var gitlabPatterns = map[string]gitlabPattern{
	"List Projects": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects$"),
	},
	"Get Project": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)$"),
	},
	"List Branches": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/repository/branches$"),
	},
	"Get Branch": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/repository/branches/[^/]+$"),
	},
	"Get Commit": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/repository/commits/[0-9a-f]{40}$"),
	},
	"Set Commit Status": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/statuses/[^/]+$"),
	},
	"Create Environment": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/environments$"),
	},
	"Stop Environment": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/environments/[0-9]+/stop$"),
	},
	"Delete Environment": {
		Method: http.MethodDelete,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/environments/[0-9]+$"),
	},
	"Get Affected Files": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/repository/compare$"),
	},
	"Get Repository Tarball": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/repository/archive$"),
	},
	"Get Spacelift Configuration": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/repository/files/[^/]*%2Espacelift%2Fconfig%2Eyml/raw$"),
	},
	"Create Deployment": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/deployments$"),
	},
	"Update Deployment": {
		Method: http.MethodPut,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/deployments/[0-9]+$"),
	},
	"List PRs": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/merge_requests$"),
	},
	"List PRs by Commit": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/repository/commits/[^/]+/merge_requests$"),
	},
	"Make PR Comment": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/merge_requests/[0-9]+/notes$"),
	},
}

func matchGitLabRequest(r *http.Request) (string, string, error) {
	for name, pattern := range gitlabPatterns {
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
