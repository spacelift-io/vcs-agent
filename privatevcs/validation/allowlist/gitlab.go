package allowlist

import (
	"net/http"
	"regexp"
)

type gitlabPattern struct {
	Method string
	Path   *regexp.Regexp
}

var gitlabPatterns = map[string]gitlabPattern{
	// The GitLab client makes a request to the API base URL to retrieve the rate limit headers
	"Get Rate Limit Info": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/?$"),
	},
	"Get Current User": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/user$"),
	},
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
	"List Environments": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/environments$"),
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
	"Get a single Merge Request": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/merge_requests/[0-9]+$"),
	},
	"Get Merge Request Approvals": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/merge_requests/[0-9]+/approvals$"),
	},
	"List Merge Requests": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/merge_requests$"),
	},
	"List Merge Requests by Commit": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/repository/commits/[^/]+/merge_requests$"),
	},
	"Make Merge Request Note": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/api/v4/projects/(?P<project>[^/]+)/merge_requests/[0-9]+/notes$"),
	},
	"Git Clone - info/refs": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile(`^/(?P<project>[^/]+\/[^/]+)\.git/info/refs$`),
	},
	"Git Clone - git-upload-pack": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile(`^/(?P<project>[^/]+\/[^/]+)\.git/git-upload-pack$`),
	},
}

func matchGitLabRequest(r *http.Request) (string, string, *string, error) {
	for name, pattern := range gitlabPatterns {
		if r.Method != pattern.Method {
			continue
		}

		if matches := pattern.Path.FindStringSubmatch(r.URL.EscapedPath()); matches != nil {
			var project string
			if index := pattern.Path.SubexpIndex("project"); index != -1 {
				project = matches[index]
			}
			return name, project, nil, nil
		}
	}
	return "", "", nil, ErrNoMatch
}
