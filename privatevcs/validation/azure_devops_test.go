package validation

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/spacelift-io/vcs-agent/nullable"
)

func TestAzureDevOpsValidation(t *testing.T) {
	testCases := []struct {
		path         string
		matches      bool
		name, method string
		project      *string
	}{
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/diffs/commits?api-version=7.1-preview.1",
			matches: true,
			name:    "Get Commit Diff",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/pullRequests/12345/threads?api-version=7.1-preview.1",
			matches: true,
			name:    "Create Pull Request Thread",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodPost,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/stats/branches?api-version=7.1-preview.1",
			matches: true,
			name:    "List Branch Stats",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/items?path=%2FREADME.md&api-version=7.1-preview.1",
			matches: true,
			name:    "Get Item",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/commits/a171eaa2c3790cbff77ec96f429e7aee5a3e724c/statuses?api-version=7.1-preview.1",
			matches: true,
			name:    "Create Commit Status",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodPost,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/pullRequests?api-version=7.1-preview.1",
			matches: true,
			name:    "List Pull Requests",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/pullRequests/1234?api-version=7.1-preview.1",
			matches: true,
			name:    "Get Pull Request",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/pullRequests/1234/labels?api-version=7.1-preview.1",
			matches: true,
			name:    "List Pull Request Labels",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories?api-version=7.1-preview.1",
			matches: true,
			name:    "List Project Repositories",
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/_apis/git/repositories?api-version=7.1-preview.1",
			matches: true,
			name:    "List Organization Repositories",
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/commits/a171eaa2c3790cbff77ec96f429e7aee5a3e724c?api-version=7.1-preview.1",
			matches: true,
			name:    "Get Commit",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/_apis",
			matches: true,
			name:    "List Resource Locations",
			method:  http.MethodOptions,
		},
		{
			path:    "/spacelift-development/_apis/ResourceAreas",
			matches: true,
			name:    "List Resource Areas",
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/pullRequests/1234/attachments/%2Fattachment.tar.gz?api-version=7.1-preview.1",
			matches: false,
			method:  http.MethodGet,
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		request, err := http.NewRequest(testCase.method, "https://github.myorg.com"+testCase.path, nil)
		require.NoError(t, err, "could not create request")

		name, project, _, err := matchAzureDevOpsRequest(request)

		if testCase.matches {
			require.NoError(t, err, "could not find match for %q", testCase.name)
			require.Equal(t, testCase.name, name)

			if testCase.project != nil {
				require.Equal(t, *testCase.project, project, "project did not match for %q", testCase.name)
			}
		} else {
			require.ErrorIs(t, ErrNoMatch, err)
		}
	}
}
