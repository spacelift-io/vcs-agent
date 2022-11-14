package allowlist

import (
	"fmt"
	"net/http"
	"regexp"
)

type bitbucketDatacenterPattern struct {
	Method string
	Path   *regexp.Regexp
}

var bitbucketDatacenterPatterns = map[string]bitbucketDatacenterPattern{
	"List Repositories": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile(`^/rest/api/1.0/repos$`),
	},
	"Get Repository": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile(`^/rest/api/1.0/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)$`),
	},
	"List Branches": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile(`^/rest/api/1.0/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)/branches$`),
	},
	"Get Commit": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile(`^/rest/api/1.0/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)/commits/(?P<commitSHA>[^/]+)$`),
	},
	"Get PR Diff": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile(`^/rest/api/1.0/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)/pull-requests/[0-9]+/diff$`),
	},
	"Set Commit Status": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile(`^/rest/api/1.0/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)/commits/(?P<commitSHA>[^/]+)/builds$`),
	},
	"Get Affected Files": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile(`^/rest/api/1.0/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)/compare/changes$`),
	},
	"Get Repository Tarball": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile(`^/rest/api/1.0/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)/archive$`),
	},
	"Get Spacelift Configuration": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile(`^/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)/raw/([^/]+/)*.spacelift/config.yml$`),
	},
	"List PRs by Branch": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile(`^/rest/api/1.0/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)/pull-requests$`),
	},
	"List PRs by Commit": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile(`^/rest/api/1.0/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)/commits/(?P<commitSHA>[^/]+)/pull-requests$`),
	},
	"Get a single Pull Request": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile(`^/rest/api/1.0/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)/pull-requests/[0-9]+$`),
	},
	"Make PR Comment": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("^/rest/api/1.0/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)/pull-requests/[0-9]+/comments$"),
	},
	"Check PR Mergeability": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/rest/api/1.0/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)/pull-requests/[0-9]+/merge$"),
	},
	"Compare Commits": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("^/rest/api/1.0/projects/(?P<projectKey>[^/]+)/repos/(?P<repositorySlug>[^/]+)/compare/commits$"),
	},
}

func matchBitbucketDatacenterRequest(r *http.Request) (string, string, error) {
	for name, pattern := range bitbucketDatacenterPatterns {
		if r.Method != pattern.Method {
			continue
		}

		if matches := pattern.Path.FindStringSubmatch(r.URL.EscapedPath()); matches != nil {
			var project string
			if index := pattern.Path.SubexpIndex("projectKey"); index != -1 {
				project = matches[index]
			}
			var repository string
			if index := pattern.Path.SubexpIndex("repositorySlug"); index != -1 {
				repository = matches[index]
			}
			var outProject string
			if repository != "" {
				outProject = fmt.Sprintf("%s/%s", project, repository)
			}
			return name, outProject, nil
		}
	}
	return "", "", ErrNoMatch
}
