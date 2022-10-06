package validation

import (
	"fmt"
	"net/http"
	"regexp"
)

type azureDevOpsPattern struct {
	Method string
	Path   *regexp.Regexp
}

var azureDevOpsPatterns = map[string]azureDevOpsPattern{
	"Get Commit Diff": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("/(?P<organization>[^/]+)/(?P<project>[^/]+)/_apis/git/repositories/(?P<repositoryId>[^/]+)/diffs/commits$"),
	},
	"Create Pull Request Thread": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("/(?P<organization>[^/]+)/(?P<project>[^/]+)/_apis/git/repositories/(?P<repositoryId>[^/]+)/pullRequests/[^/]+/threads$"),
	},
	"List Branch Stats": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("/(?P<organization>[^/]+)/(?P<project>[^/]+)/_apis/git/repositories/(?P<repositoryId>[^/]+)/stats/branches$"),
	},
	"Get Item": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("/(?P<organization>[^/]+)/(?P<project>[^/]+)/_apis/git/repositories/(?P<repositoryId>[^/]+)/items$"),
	},
	"Create Commit Status": {
		Method: http.MethodPost,
		Path:   regexp.MustCompile("/(?P<organization>[^/]+)/(?P<project>[^/]+)/_apis/git/repositories/(?P<repositoryId>[^/]+)/commits/[^/]+/statuses$"),
	},
	"List Pull Requests": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("/(?P<organization>[^/]+)/(?P<project>[^/]+)/_apis/git/repositories/(?P<repositoryId>[^/]+)/pullRequests$"),
	},
	"Get Pull Request": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("/(?P<organization>[^/]+)/(?P<project>[^/]+)/_apis/git/repositories/(?P<repositoryId>[^/]+)/pullRequests/[^/]+$"),
	},
	"List Pull Request Labels": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("/(?P<organization>[^/]+)/(?P<project>[^/]+)/_apis/git/repositories/(?P<repositoryId>[^/]+)/pullRequests/[^/]+/labels$"),
	},
	"List Repositories": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("(/(?P<organization>[^/]+))?/(?P<project>[^/]+)/_apis/git/repositories$"),
	},
	"Get Commit": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("/(?P<organization>[^/]+)/(?P<project>[^/]+)/_apis/git/repositories/(?P<repositoryId>[^/]+)/commits/[^/]+$"),
	},
	"List Resource Locations": {
		Method: http.MethodOptions,
		Path:   regexp.MustCompile("/(?P<organization>[^/]+)/_apis$"),
	},
	"List Resource Areas": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("/(?P<organization>[^/]+)/_apis/ResourceAreas$"),
	},
	"List Policy Evaluations": {
		Method: http.MethodGet,
		Path:   regexp.MustCompile("/(?P<organization>[^/]+)/(?P<project>[^/]+)/_apis/policy/evaluations$"),
	},
}

func matchAzureDevOpsRequest(r *http.Request) (string, string, *string, error) {
	for name, pattern := range azureDevOpsPatterns {
		if r.Method != pattern.Method {
			continue
		}

		if matches := pattern.Path.FindStringSubmatch(r.URL.EscapedPath()); matches != nil {
			var organization, project, repositoryID string
			if index := pattern.Path.SubexpIndex("organization"); index != -1 {
				organization = matches[index]
			}
			if index := pattern.Path.SubexpIndex("project"); index != -1 {
				project = matches[index]
			}
			if index := pattern.Path.SubexpIndex("repositoryId"); index != -1 {
				repositoryID = matches[index]
			}

			var projectName string
			if organization != "" && project != "" && repositoryID != "" {
				projectName = fmt.Sprintf("%s/%s/%s", organization, project, repositoryID)
			}

			return name, projectName, nil, nil
		}
	}

	// Temporarily just allow any resources - we'll restrict later once we're certain we have
	// a stable set of endpoints in the validation rules
	return "Unknown Request", "", nil, nil
}
